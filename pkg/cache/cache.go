// SPDX-FileCopyrightText: 2022 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"fmt"
	"sync"
	"time"
)

type Cache struct {
	stop chan bool

	mutex           sync.RWMutex
	wg              sync.WaitGroup
	cache           map[string]cachedObject
	cachedObjectTTL int64
}

type cachedObject struct {
	bytes               []byte
	expirationTimestamp int64
}

type Cacher interface {
	Get(key string) []byte
	Update(key string, bytes []byte)
	StopRefresh()
}

func NewCache(refreshInterval time.Duration, cachedObjectTTL int64) (*Cache, error) {
	refreshIntervalInSeconds := int64(refreshInterval.Seconds())
	if refreshIntervalInSeconds > cachedObjectTTL {
		return nil, fmt.Errorf("the refresh interval of %v seconds should not be greater than the cached object validity duration seconds", refreshIntervalInSeconds)
	}
	dc := &Cache{
		stop:            make(chan bool),
		cache:           make(map[string]cachedObject),
		cachedObjectTTL: cachedObjectTTL,
	}

	dc.wg.Add(1)
	go func(refreshInterval time.Duration) {
		defer dc.wg.Done()
		dc.refreshLoop(refreshInterval)
	}(refreshInterval)

	return dc, nil
}

func (dc *Cache) refreshLoop(refreshInterval time.Duration) {
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case quit := <-dc.stop:
			if quit {
				return
			}
		case <-ticker.C:
			dc.mutex.Lock()
			for k, v := range dc.cache {
				if v.expirationTimestamp <= time.Now().Unix() {
					delete(dc.cache, k)
				}
			}
			dc.mutex.Unlock()
		}
	}
}

func (dc *Cache) Get(key string) []byte {
	dc.mutex.RLock()
	defer dc.mutex.RUnlock()
	if v, ok := dc.cache[key]; ok {
		return v.bytes
	}

	return nil
}

func (dc *Cache) Update(key string, bytes []byte) {
	dc.mutex.Lock()
	defer dc.mutex.Unlock()
	dc.cache[key] = cachedObject{
		expirationTimestamp: time.Now().Unix() + dc.cachedObjectTTL,
		bytes:               bytes,
	}
}

func (dc *Cache) StopRefresh() {
	dc.stop <- true
	close(dc.stop)
	dc.wg.Wait()
}
