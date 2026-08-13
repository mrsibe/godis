package database

import "sync"

// DB 是并发安全的字符串键值存储，所有方法都可被多个 goroutine 同时调用。
type DB struct {
	mu   sync.RWMutex
	data map[string]string
}

// New 创建并返回一个空的 DB。
func New() *DB {
	return &DB{
		data: make(map[string]string),
	}
}

// Set 将 key 映射到 value，覆盖已存在的值。
func (db *DB) Set(key, value string) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.data[key] = value
}

// Get 返回 key 对应的值；key 不存在时返回零值和 false。
func (db *DB) Get(key string) (string, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	value, ok := db.data[key]
	return value, ok
}

// Del 删除 key，返回是否真的删除了；key 不存在时返回 false。
func (db *DB) Del(key string) bool {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, ok := db.data[key]; !ok {
		return false
	}

	delete(db.data, key)
	return true
}
