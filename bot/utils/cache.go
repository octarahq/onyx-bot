package utils

import (
	"sync"
	"time"
)

type CacheItem struct {
	Value      interface{}
	Expiration int64
}

type GlobalCache struct {
	Mu    sync.RWMutex
	Items map[string]CacheItem
}

var Cache = &GlobalCache{
	Items: make(map[string]CacheItem),
}

func (c *GlobalCache) Set(key string, value interface{}, duration time.Duration) {
	var exp int64
	if duration > 0 {
		exp = time.Now().Add(duration).UnixNano()
	}

	c.Mu.Lock()
	defer c.Mu.Unlock()

	c.Items[key] = CacheItem{
		Value:      value,
		Expiration: exp,
	}
}

func (c *GlobalCache) Get(key string) (interface{}, bool) {
	c.Mu.RLock()
	defer c.Mu.RUnlock()

	item, found := c.Items[key]
	if !found {
		return nil, false
	}

	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return nil, false
	}

	return item.Value, true
}

func (c *GlobalCache) Delete(key string) {
	c.Mu.Lock()
	defer c.Mu.Unlock()
	delete(c.Items, key)
}

func (c *GlobalCache) Cleanup() {
	c.Mu.Lock()
	defer c.Mu.Unlock()

	now := time.Now().UnixNano()
	for k, v := range c.Items {
		if v.Expiration > 0 && now > v.Expiration {
			delete(c.Items, k)
		}
	}
}

func init() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			Cache.Cleanup()
		}
	}()
}
