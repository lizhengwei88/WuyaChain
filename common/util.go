package common

import lru "github.com/hashicorp/golang-lru"

// MustNewCache creates a LRU cache with specified size. Panics on any error.
func MustNewCache(size int) *lru.Cache {
	cache, err := lru.New(size)
	if err != nil {
		panic(err) // error occurs only when size <= 0.
	}

	return cache
}