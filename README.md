# Go Supercache

A double cache system (memory and redis) to fast and distribuited cache.

## Install

```bash
$ go get github.com/alboompro/fastcache
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/alboompro/fastcache"
)

// Object example to save into
type Object struct {
	ID   string
	Name string
}

func main() {
	// Creating cache instance
	cache := fastcache.New(&fastcache.CacheNewConfig{
		Counters:         3,   // Parallels memory cache
		TTLCacheSec:      300, // Redis cache TTL
		TTLRevalidateSec: 30,  // Revalidate and Recache TTL
	})

	// Retrieve with set cache
	var obj = new(Object)
	err := cache.Get(&fastcache.ItemRequest{
		Key:   "example-identifier",
		Value: obj, // destination,
		Do: func(req *fastcache.ItemRequest) (interface{}, error) {
			// make your code to retrive the original data
			return &Object{
				ID:   "1",
				Name: "Named",
			}, nil
		},
		TTL:           300, // override redis cache original TTL, optional
		TTLRevalidate: 30,  // override original revalidate TTL, optional
	})

	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", obj) // print object into your terminal, &{ID:1 Name:Named}
}
```

### Full configuration example

Cache config:

```go

var config = fastcache.CacheNewConfig{
    Counters:         3,   // Parallels memory cache
    TTLCacheSec:      300, // Redis cache TTL
    TTLRevalidateSec: 30,  // Revalidate and Recache TTL

    // Memory cache configuration
    // see: https://pkg.go.dev/github.com/dgraph-io/ristretto@v0.1.0#Config
    CacheConfig: &ristretto.Config{
        MaxCost: 1 << 30, // default value
        BufferItems: 1e3, // default value
        NumCounters: 1e4, // default value
    },

    // Memory cache configuration
    // see: https://pkg.go.dev/github.com/go-redis/redis/v8@v8.11.5#Ring
    RedisConfig: &redis.RingOptions{
        Addrs: map[string]string{
			"localhost": ":6379",
		}, // default value
    },

	// RedisMode store the mode of redis connection
	// options: fastcache.MODE_SINGLE OR fastcache.MODE_CLUSTER
	RedisMode: fastcache.MODE_SINGLE, // default value
}

var req = fastcache.ItemRequest{
    Key:   "example-identifier",
    Value: obj, // destination,
    Do: func(req *fastcache.ItemRequest) (interface{}, error) {
        // make your code to retrive the original data
        return &Object{
            ID:   "1",
            Name: "Named",
        }, nil
    },
    TTL:           300, // override redis cache original TTL, optional
    TTLRevalidate: 30,  // override original revalidate TTL, optional
}

```

## Development

Requires redis server running on `localhost:6379`. Or run:

```bash
$ docker-compose up -d
```
