package common

import (
	lru "github.com/hashicorp/golang-lru"
	"math/rand"
	"reflect"
)

// Bytes is a array byte that is converted to hex string display format when marshal
type Bytes []byte

// MustNewCache creates a LRU cache with specified size. Panics on any error.
func MustNewCache(size int) *lru.Cache {
	cache, err := lru.New(size)
	if err != nil {
		panic(err) // error occurs only when size <= 0.
	}

	return cache
}

// CopyBytes copies and returns a new bytes from the specified source bytes.
func CopyBytes(src []byte) []byte {
	if src == nil {
		return nil
	}

	dest := make([]byte, len(src))
	copy(dest, src)
	return dest
}

// Shuffle shuffles items in slice
func Shuffle(slice interface{}) {
	rv := reflect.ValueOf(slice)
	swap := reflect.Swapper(slice)
	length := rv.Len()
	for i := length - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		swap(i, j)
	}
}
