package database

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
)

// strObj 构造一个字符串类型的 Object，方便测试。
func strObj(s string) *Object {
	return &Object{Type: StringType, Value: s}
}

func TestSetGet(t *testing.T) {
	db := New()

	db.Set("name", strObj("godis"))

	got, ok := db.Get("name")
	if !ok {
		t.Fatal("Get returned ok=false, want true")
	}
	if got.Value != "godis" {
		t.Fatalf("Get returned %q, want %q", got.Value, "godis")
	}
}

func TestGetMissingKey(t *testing.T) {
	db := New()

	got, ok := db.Get("nope")
	if ok {
		t.Fatalf("Get returned ok=true, want false")
	}
	if got != nil {
		t.Fatalf("Get returned %v, want nil", got)
	}
}

func TestSetOverwrites(t *testing.T) {
	db := New()

	db.Set("key", strObj("first"))
	db.Set("key", strObj("second"))

	got, ok := db.Get("key")
	if !ok {
		t.Fatal("Get returned ok=false, want true")
	}
	if got.Value != "second" {
		t.Fatalf("Get returned %q, want %q", got.Value, "second")
	}
}

func TestDel(t *testing.T) {
	db := New()

	db.Set("key", strObj("value"))

	if ok := db.Del("key"); !ok {
		t.Fatal("Del returned false, want true")
	}

	_, ok := db.Get("key")
	if ok {
		t.Fatal("Get after Del returned ok=true, want false")
	}
}

func TestDelMissingKey(t *testing.T) {
	db := New()

	if ok := db.Del("nope"); ok {
		t.Fatal("Del returned true, want false")
	}
}

func TestEmptyKeyValue(t *testing.T) {
	db := New()

	// 空 key 和空 value 都应该能正常存取
	db.Set("", strObj("value"))
	got, ok := db.Get("")
	if !ok || got.Value != "value" {
		t.Fatalf("Get(empty key) = (%q, %v), want (%q, true)", got.Value, ok, "value")
	}

	db.Set("key", strObj(""))
	got, ok = db.Get("key")
	if !ok || got.Value != "" {
		t.Fatalf("Get(empty value) = (%q, %v), want (%q, true)", got.Value, ok, "")
	}
}

func TestManyKeys(t *testing.T) {
	db := New()

	const n = 1000
	for i := range n {
		db.Set(strconv.Itoa(i), strObj(strconv.Itoa(i*2)))
	}

	for i := range n {
		want := strconv.Itoa(i * 2)
		got, ok := db.Get(strconv.Itoa(i))
		if !ok {
			t.Fatalf("Get(%d) returned ok=false, want true", i)
		}
		if got.Value != want {
			t.Fatalf("Get(%d) = %q, want %q", i, got.Value, want)
		}
	}
}

// TestConcurrent 用多个 goroutine 并发读写，配合 go test -race 验证并发安全。
func TestConcurrent(t *testing.T) {
	db := New()

	const goroutines = 32
	const opsPerGoroutine = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := range goroutines {
		go func(g int) {
			defer wg.Done()
			prefix := fmt.Sprintf("g%d", g)
			for i := range opsPerGoroutine {
				key := fmt.Sprintf("%s:%d", prefix, i)
				db.Set(key, strObj(key))

				got, ok := db.Get(key)
				if !ok || got.Value != key {
					t.Errorf("Get(%q) = (%q, %v), want (%q, true)", key, got.Value, ok, key)
				}

				// 删除一半的 key，模拟真实读写混合
				if i%2 == 0 {
					if ok := db.Del(key); !ok {
						t.Errorf("Del(%q) returned false, want true", key)
					}
					if _, ok := db.Get(key); ok {
						t.Errorf("Get(%q) after Del returned ok=true, want false", key)
					}
				}
			}
		}(g)
	}

	wg.Wait()

	// 验证残留数据正确：奇数 key 应保留，偶数 key 应已删除
	for g := range goroutines {
		prefix := fmt.Sprintf("g%d", g)
		for i := range opsPerGoroutine {
			key := fmt.Sprintf("%s:%d", prefix, i)
			_, ok := db.Get(key)
			want := i%2 == 1
			if ok != want {
				t.Errorf("Get(%q) ok=%v, want %v", key, ok, want)
			}
		}
	}
}
