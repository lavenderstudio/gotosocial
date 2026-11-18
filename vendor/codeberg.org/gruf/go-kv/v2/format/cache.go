package format

import (
	"sync/atomic"
	"unsafe"
)

// cache is a concurrency-safe map[K]V cache, designed for
// heavy reads but with unfortunately expensive writes. it
// is designed such that after some initial load period in
// which functions are cached by types, all future ops are reads.
type cache[K comparable, V any] struct{ p unsafe.Pointer }

// Get will check cache for format func under key.
func (c *cache[K, V]) Get(k K) (v V, ok bool) {
	if p := c.load(); p != nil {
		v, ok = (*p)[k]
	}
	return
}

// Put will place given value in cache
// under key, if not already exists.
func (c *cache[K, V]) Put(k K, v V) {
	for {
		p := c.load()

		var cache map[K]V

		if p != nil {
			if _, ok := (*p)[k]; ok {
				return
			}

			cache = make(map[K]V, len(*p)+1)
			for key, value := range *p {
				cache[key] = value
			}
		} else {
			cache = make(map[K]V, 1)
		}

		cache[k] = v

		if c.cas(p, &cache) {
			return
		}
	}
}

// load is a typed wrapper around atomic.LoadPointer().
func (c *cache[K, V]) load() *map[K]V {
	return (*map[K]V)(atomic.LoadPointer(&c.p))
}

// cas is a typed wrapper around atomic.CompareAndSwapPointer().
func (c *cache[K, V]) cas(old, new *map[K]V) bool {
	return atomic.CompareAndSwapPointer(&c.p, unsafe.Pointer(old), unsafe.Pointer(new))
}
