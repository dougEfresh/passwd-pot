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
	log "github.com/Sirupsen/logrus"
	"github.com/orcaman/concurrent-map"
	"time"
)

// Cache is a synchronised map of items that auto-expire once stale
type Cache struct {
	cm cmap.ConcurrentMap
}

func (cache *Cache) set(key string, data int64) {
	cache.cm.Set(key, data)
}

func (cache *Cache) get(key string) (data int64, found bool) {
	if cache.cm.Has(key) {
		v, found := cache.cm.Get(key)
		return v.(int64), found
	}
	return 0, false
}

func (cache *Cache) Delete(key string) {
	cache.cm.Remove(key)
}

func (cache *Cache) Clear() {
	for item := range cache.cm.IterBuffered() {
		log.Infof("Removing cached item %s", item.Key)
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
