package fastcache

import (
	"time"

	rcache "github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

type Object struct {
	Str string
	Num int
}

type RedisMODE string

var (
	MODE_CLUSTER = RedisMODE("cluster")
	MODE_SINGLE  = RedisMODE("single")
)

func newRedisConn(config *CacheNewConfig) *rcache.Cache {
	var mycache *rcache.Cache
	if config.RedisMode == MODE_CLUSTER {
		var addrs []string
		for host, port := range config.RedisConfig.Addrs {
			addrs = append(addrs, host+port)
		}
		rdb := redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: addrs,
		})
		mycache = rcache.New(&rcache.Options{
			Redis:      rdb,
			LocalCache: rcache.NewTinyLFU(int(config.TTLCacheSec), time.Second),
		})
	} else {
		mycache = rcache.New(&rcache.Options{
			Redis:      redis.NewRing(config.RedisConfig),
			LocalCache: rcache.NewTinyLFU(int(config.TTLCacheSec), time.Second),
		})
	}

	return mycache
}

func (c *Cache) redisLoad(req *ItemRequest) (*entry, error) {
	var e = new(entry)
	if err := c.redis_conn.Once(&rcache.Item{
		Key:   req.Key,
		Value: e, // destination
		Do: func(*rcache.Item) (interface{}, error) {
			if newEntry, err := req.loadEntry(c); err == nil {
				return newEntry, nil
			} else {
				return nil, err
			}
		},
		TTL:            req.expiresAt(c),
		SkipLocalCache: true,
	}); err != nil {
		return nil, err
	}

	return e, nil
}
