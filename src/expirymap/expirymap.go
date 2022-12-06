package expirymap

import (
	"sync"
	"time"
)

type ExpiryMap struct {
	// The internal map that stores the key-value pairs
	m map[string]string

	// A mutex that protects access to the map
	mutex sync.RWMutex

	// A map that stores the expiration time for each key-value pair
	expiryMap map[string]time.Time
}

func New() ExpiryMap {
	return ExpiryMap{
		m:         make(map[string]string),
		expiryMap: make(map[string]time.Time),
	}
}

func (em *ExpiryMap) Set(key string, value string, expiry time.Duration) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	em.m[key] = value
	em.expiryMap[key] = time.Now().Add(expiry)
}

func (em *ExpiryMap) Get(key string) (string, bool) {
	em.mutex.RLock()
	defer em.mutex.RUnlock()

	if value, ok := em.m[key]; ok {
		expiry := em.expiryMap[key]
		if time.Now().Before(expiry) {
			return value, true
		}

		delete(em.m, key)
		delete(em.expiryMap, key)
	}

	return "", false
}

func (em *ExpiryMap) Delete(key string) {
	em.mutex.Lock()
	defer em.mutex.Unlock()

	delete(em.m, key)
	delete(em.expiryMap, key)
}
