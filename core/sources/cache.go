package sources

import (
	"time"

	"github.com/dgraph-io/ristretto"
)

type (
	Cache interface {
		Get(key interface{}) (interface{}, bool)
		Set(key, value interface{}, cost int64) bool
		SetWithTTL(key, value interface{}, cost int64, ttl time.Duration) bool
		Del(key interface{})
		Clear()
		Wait()
	}

	cache struct {
		c *ristretto.Cache
	}
)

func NewCache(maxCost, numCounters, bufferItems int64) Cache {
	c, err := ristretto.NewCache(&ristretto.Config{
		MaxCost:     maxCost,
		NumCounters: numCounters,
		BufferItems: bufferItems,
	})
	if err != nil {
		panic(err)
	}

	return &cache{c: c}
}

func (c *cache) Get(key interface{}) (interface{}, bool) {
	return c.c.Get(key)
}

func (c *cache) Set(key, value interface{}, cost int64) bool {
	return c.c.Set(key, value, cost)
}

func (c *cache) SetWithTTL(key, value interface{}, cost int64, ttl time.Duration) bool {
	if ttl == 0 {
		ttl = 24 * time.Hour
	}

	return c.c.SetWithTTL(key, value, cost, ttl)
}

func (c *cache) Del(key interface{}) {
	c.c.Del(key)
}

func (c *cache) Clear() {
	c.c.Clear()
}

func (c *cache) Wait() {
	c.c.Wait()
}
