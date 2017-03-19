// Copyright © 2017 Douglas Chimento <dchimento@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"time"
)

// Cache is a synchronised map of items that auto-expire once stale
type Cache struct {
	ttl   time.Duration
	items ConcurrentMap
}

// Set is a thread-safe way to add new items to the map
func (cache *Cache) Set(key string, data int64) {
	item := Item{data: data}
	item.touch(cache.ttl)
	cache.items.Set(key, item)
}

// Get is a thread-safe way to lookup items
// Every lookup, also touches the item, hence extending it's life
func (cache *Cache) Get(key string) (data int64, found bool) {
	item, exists := cache.items.Get(key)
	if !exists || item.expired() {
		data = 0
		found = false
	} else {
		data = item.data
		found = true
	}
	return
}

// Count returns the number of items in the cache
// (helpful for tracking memory leaks)
func (cache *Cache) Count() int {
	count := len(cache.items)
	return count
}

func (cache *Cache) cleanup() {
	for key, item := range cache.items.Items() {
		if item.expired() {
			cache.items.Remove(key)
		}
	}
}

func (cache *Cache) startCleanupTimer() {
	duration := cache.ttl
	if duration < time.Second {
		duration = time.Second
	}
	ticker := time.Tick(duration)
	go (func() {
		for {
			select {
			case <-ticker:
				cache.cleanup()
			}
		}
	})()
}

// NewCache is a helper to create instance of the Cache struct
func NewCache(duration time.Duration) *Cache {
	cache := &Cache{
		ttl:   duration,
		items: NewConcurrentMap(),
	}
	cache.startCleanupTimer()
	return cache
}
