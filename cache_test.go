package fastcache

import (
	"fmt"
	"testing"
	"time"
)

type authItem struct {
	ID   string
	Name string
}

func validConfig() *CacheNewConfig {
	return &CacheNewConfig{
		Counters:         1,
		TTLCacheSec:      5,
		TTLRevalidateSec: 1,
	}
}

func item(obj interface{}, key string, rets ...string) *ItemRequest {
	return &ItemRequest{
		Key:   key,
		Value: obj,
		Do: func(req *ItemRequest) (interface{}, error) {
			r := &authItem{
				ID:   "123",
				Name: "Named",
			}

			for i, ret := range rets {
				if i == 0 {
					r.ID = ret
				} else {
					r.Name = ret
				}
			}

			return r, nil
		},
	}
}

func TestCacheCreation(t *testing.T) {
	cache := New(validConfig())
	if cache == nil {
		t.Error("cache is nil")
	}
}

func Test_CanAddEntryToCache(t *testing.T) {
	cc := New(validConfig())
	var obj = new(authItem)

	err := cc.Get(item(obj, "Test_CanAddEntryToCache"))
	if err != nil {
		t.Error(err)
	}

	if obj.ID != "123" {
		fmt.Printf("%+v\n", obj)
		t.Errorf("ID is not 123, but %s", obj.ID)
	}
}

func Test_CanRetrieveCacheMemoryEntry(t *testing.T) {
	cc := New(validConfig())
	var obj = new(authItem)
	var obj2 = new(authItem)

	cc.Get(item(obj, "Test_CanRetrieveCacheMemoryEntry", "1"))
	cc.Get(item(obj2, "Test_CanRetrieveCacheMemoryEntry", "11"))
	if obj2.ID != obj.ID {
		t.Errorf("ID is not 1, but %s", obj2.ID)
	}
}

func Test_CanRetrieveExpiredCache(t *testing.T) {
	cc := New(&CacheNewConfig{
		Counters:         1,
		TTLCacheSec:      3,
		TTLRevalidateSec: 1,
	})
	// cc.redis_conn.Get()
	var obj = new(authItem)
	var obj2 = new(authItem)
	var obj3 = new(authItem)

	// Create a new cache entry with id 1
	cc.Get(item(obj, "Test_CanRetrieveExpiredCache", "1"))
	// wait time to active revalidator
	time.Sleep(2 * time.Second)
	// Get cache item with revalidation, should return
	// id 1 and save new cache with id 11
	cc.Get(item(obj2, "Test_CanRetrieveExpiredCache", "11"))
	if obj2.ID != obj.ID {
		t.Errorf("obj2.ID is not 1, but %s", obj2.ID)
	}
	// Retrieve cache item, should return id 11
	time.Sleep(500 * time.Millisecond)
	cc.Get(item(obj3, "Test_CanRetrieveExpiredCache", "11"))
	if obj3.ID == obj.ID {
		t.Errorf("obj3.ID is not 11, but %s", obj3.ID)
	}
}

func Test_CustomTTLInRequest(t *testing.T) {
	cc := New(&CacheNewConfig{
		Counters:         1,
		TTLCacheSec:      3,
		TTLRevalidateSec: 1,
	})
	// cc.redis_conn.Get()
	var obj = new(authItem)
	var obj2 = new(authItem)
	var obj3 = new(authItem)

	// Create a new cache entry with id 1
	i1 := item(obj, "Test_CustomTTLInRequest", "1")
	i1.TTL = 60
	i1.TTLRevalidate = 30
	cc.Get(i1)
	// wait time to active revalidator
	time.Sleep(2 * time.Second)
	// Get cache item with revalidation, should return
	// id 1 and save new cache with id 11
	cc.Get(item(obj2, "Test_CustomTTLInRequest", "11"))
	if obj2.ID != obj.ID {
		t.Errorf("obj2.ID is not 1, but %s", obj2.ID)
	}
	// Retrieve cache item, should return id 11
	time.Sleep(500 * time.Millisecond)
	cc.Get(item(obj3, "Test_CustomTTLInRequest", "11"))
	if obj3.ID != obj.ID {
		t.Errorf("obj3.ID is not 1, but %s", obj3.ID)
	}
}
