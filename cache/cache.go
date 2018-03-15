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

package cache

import (
	"time"

	"github.com/orcaman/concurrent-map"
)

// Cache is a synchronised map of items that auto-expire once stale
type Cache struct {
	cm cmap.ConcurrentMap
}

// Set cache
func (cache *Cache) Set(key string, data int64) {
	cache.cm.Set(key, data)
}

// Get cache
func (cache *Cache) Get(key string) (int64, bool) {
	if cache.cm.Has(key) {
		v, _ := cache.cm.Get(key)
		return v.(int64), true
	}
	return 0, false
}

// Delete the cache
func (cache *Cache) Delete(key string) {
	cache.cm.Remove(key)
}

// Clear the cache
func (cache *Cache) Clear() {
	for item := range cache.cm.IterBuffered() {
		cache.cm.Remove(item.Key)
	}
}

func (cache *Cache) startCleanupTimer() {
	for {
		time.Sleep(48 * time.Hour)
		cache.Clear()
	}
}

// NewCache is a helper to create instance of the Cache struct
func NewCache() *Cache {
	cache := &Cache{
		cm: cmap.New(),
	}
	go cache.startCleanupTimer()
	return cache
}
