package database

import "sync"

type DB struct {
	mu   sync.RWMutex
	data map[string]*Object
}

func New() *DB {
	return &DB{
		data: make(map[string]*Object),
	}
}

func (db *DB) Set(key string, value *Object) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.data[key] = value
}

func (db *DB) Get(key string) (*Object, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	value, ok := db.data[key]
	return value, ok
}

func (db *DB) Del(key string) bool {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.data[key]; !ok {
		return false
	}

	delete(db.data, key)
	return true
}
