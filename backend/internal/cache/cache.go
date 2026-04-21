package cache

import (
	"strings"
	"sync"
	"time"
)

type entry struct {
	value     any
	expiresAt time.Time
}

// Cache is a tiny TTL store. Good enough for a single-node, low-traffic OJ;
// replace with Redis if the app ever needs multi-process coherency.
type Cache struct {
	m sync.Map // key string -> entry
}

func New() *Cache { return &Cache{} }

func (c *Cache) Get(key string) (any, bool) {
	v, ok := c.m.Load(key)
	if !ok {
		return nil, false
	}
	e := v.(entry)
	if time.Now().After(e.expiresAt) {
		c.m.Delete(key)
		return nil, false
	}
	return e.value, true
}

func (c *Cache) Set(key string, value any, ttl time.Duration) {
	c.m.Store(key, entry{value: value, expiresAt: time.Now().Add(ttl)})
}

// Invalidate drops any key starting with prefix.
func (c *Cache) Invalidate(prefix string) {
	c.m.Range(func(k, _ any) bool {
		if ks, ok := k.(string); ok && strings.HasPrefix(ks, prefix) {
			c.m.Delete(k)
		}
		return true
	})
}
