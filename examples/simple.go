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
