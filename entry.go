package fastcache

import (
	"time"

	"github.com/dgraph-io/ristretto"
	rcache "github.com/go-redis/cache/v8"
)

type entry struct {
	RevalidateAt time.Time
	ExpiresAt    time.Time
	Key          string
	Value        interface{}
}

func (e *entry) revalid(c *Cache, r *ristretto.Cache, i *ItemRequest) {
	if time.Now().After(e.RevalidateAt) {
		go func() {
			if newEntry, err := i.loadEntry(c); err == nil {
				r.SetWithTTL(i.Key, newEntry, 1, i.expiresAt(c))

				// wait for value to pass through buffers
				r.Wait()

				c.redis_conn.Set(&rcache.Item{
					Key:            i.Key,
					Value:          newEntry,
					SkipLocalCache: true,
					TTL:            i.expiresAt(c),
				})
			}
		}()
	}
}

func (e *entry) expired() bool {
	return time.Now().After(e.ExpiresAt)
}

func (i *ItemRequest) loadEntry(c *Cache) (*entry, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if result, err := i.Do(i); err == nil {
		return &entry{
			RevalidateAt: time.Now().Add(i.revalidateAt(c)),
			ExpiresAt:    time.Now().Add(i.expiresAt(c)),
			Key:          i.Key,
			Value:        result,
		}, nil
	} else {
		return nil, err
	}
}
