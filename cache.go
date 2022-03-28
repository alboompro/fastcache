package fastcache

import (
	"hash/crc32"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"
	rcache "github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

var (
	ErrKeyEmpty        = errors.New("key is empty")
	ErrKeyNotSupported = errors.New("key is not supported")
)

// Cache returns the underlying cache.
type Cache struct {
	ristretto_instances []*ristretto.Cache
	mus                 []*sync.Mutex
	mu                  *sync.Mutex
	redis_conn          *rcache.Cache
	counter             int8
	config              *CacheNewConfig
}

// CacheNewConfig is a struct that holds the configuration for the new cache
type CacheNewConfig struct {
	// Number of paparells faster memory cache
	Counters int8

	// TTL in seconds to redis cache
	TTLCacheSec int32

	// Time in seconds to wait before trying to revalidate a cache entry
	TTLRevalidateSec int

	// Ristretto memory cache config
	CacheConfig *ristretto.Config

	// Redis connection config
	RedisConfig *redis.RingOptions
}

// ItemRequest is a struct that holds the request for the cache
type ItemRequest struct {
	// Key is the key to the cache naming
	Key   string
	Value interface{}

	// Do returns value to be cached.
	Do func(request *ItemRequest) (interface{}, error)

	// TTL is the cache expiration time.
	// Default TTL is cache configuration value.
	TTL int

	// TTLRevalidate is the memory cache expiration time.
	// Default TTL is cache configuration value.
	TTLRevalidate int
}

func (i *ItemRequest) revalidateAt(c *Cache) time.Duration {
	if i.TTLRevalidate > 0 {
		return time.Duration(i.TTLRevalidate) * time.Second
	}
	return time.Duration(c.config.TTLRevalidateSec) * time.Second
}

func (i *ItemRequest) expiresAt(c *Cache) time.Duration {
	if i.TTL > 0 {
		return time.Duration(i.TTL) * time.Second
	}
	return time.Duration(c.config.TTLCacheSec) * time.Second
}

// NewCache creates a new cache
func New(config *CacheNewConfig) *Cache {
	c := &Cache{
		config:  config,
		mu:      &sync.Mutex{},
		counter: 0,
	}

	if c.config.RedisConfig == nil {
		c.config.RedisConfig = &redis.RingOptions{}
	}

	if (c.config.RedisConfig.Addrs == nil) || (len(c.config.RedisConfig.Addrs) == 0) {
		c.config.RedisConfig.Addrs = map[string]string{
			"localhost": ":6379",
		}
	}
	c.redis_conn = newRedisConn(c.config)

	if c.config.CacheConfig == nil {
		c.config.CacheConfig = &ristretto.Config{}
	}

	if c.config.CacheConfig.MaxCost == 0 {
		c.config.CacheConfig.MaxCost = 1 << 30
	}

	if c.config.CacheConfig.BufferItems == 0 {
		c.config.CacheConfig.BufferItems = 1e3
	}

	if c.config.CacheConfig.NumCounters == 0 {
		c.config.CacheConfig.NumCounters = 1e4
	}

	for i := 0; i < int(config.Counters); i++ {
		rist, err := ristretto.NewCache(c.config.CacheConfig)
		if err != nil {
			panic(err)
		}

		c.mus = append(c.mus, &sync.Mutex{})
		c.ristretto_instances = append(c.ristretto_instances, rist)
		c.counter += 1
	}

	return c
}

// Get gets the ItemRequest.Value for the given ItemRequest.Key from the
// cache or executes, caches, and returns the results of the given ItemRequest.Do
// making sure that only one execution is in-flight for a given item.Key
// at a time. If a duplicate comes in, the duplicate caller waits for the
// original to complete and receives the same results. If the revalidation
// TTL is expired returns the same value and execute the ItemRequest.Do to
// store the new value in the cache.
func (c *Cache) Get(req *ItemRequest) error {
	if req.Key == "" {
		return ErrKeyEmpty
	}

	idx := int(crc32.ChecksumIEEE([]byte(req.Key)) % uint32(c.counter))

	c.mus[idx].Lock()
	defer c.mus[idx].Unlock()

	ris := c.ristretto_instances[idx]

	ristObj, exists := ris.Get(req.Key)
	if exists {
		if e, ok := ristObj.(*entry); ok {
			e.revalid(c, ris, req)
			if !e.expired() {
				mapstructure.Decode(e.Value, req.Value)
				return nil
			}
		}
	}

	if e, err := c.redisLoad(req); err == nil {
		e.revalid(c, ris, req)
		mapstructure.Decode(e.Value, req.Value)
		// fmt.Printf("redis entry: %+v\n", req.Value)
		ris.SetWithTTL(req.Key, e, 1, req.expiresAt(c))

		// wait for value to pass through buffers
		ris.Wait()

		// req.Value = e.Value
		return nil
	} else {
		return err
	}
}
