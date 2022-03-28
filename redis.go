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

func newRedisConn(config *CacheNewConfig) *rcache.Cache {
	ring := redis.NewRing(config.RedisConfig)

	mycache := rcache.New(&rcache.Options{
		Redis:      ring,
		LocalCache: rcache.NewTinyLFU(int(config.TTLCacheSec), time.Second),
	})
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
