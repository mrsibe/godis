package command

import (
	"reflect"
	"testing"

	"github.com/mrsibe/godis/internal/database"
	"github.com/mrsibe/godis/internal/resp"
)

func bulk(s string) resp.Value { return resp.Value{Type: resp.BulkString, String: s} }

func TestSet(t *testing.T) {
	db := database.New()

	got := Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{
		bulk("SET"), bulk("name"), bulk("godis"),
	}})
	if want := (resp.Value{Type: resp.SimpleString, String: "OK"}); !reflect.DeepEqual(got, want) {
		t.Errorf("SET = %+v, want %+v", got, want)
	}

	// 验证真的写进去了
	value, ok := db.Get("name")
	if !ok || value != "godis" {
		t.Errorf("db.Get(\"name\") = (%q, %v), want (\"godis\", true)", value, ok)
	}
}

func TestSetWrongArgs(t *testing.T) {
	db := database.New()

	// 缺参数
	got := Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{bulk("SET"), bulk("key")}})
	if got.Type != resp.Error {
		t.Errorf("SET 缺参数 = %+v, want Error", got)
	}

	// 多参数
	got = Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{
		bulk("SET"), bulk("key"), bulk("v1"), bulk("v2"),
	}})
	if got.Type != resp.Error {
		t.Errorf("SET 多参数 = %+v, want Error", got)
	}
}

func TestGet(t *testing.T) {
	db := database.New()
	db.Set("name", "godis")

	got := Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{
		bulk("GET"), bulk("name"),
	}})
	if want := (resp.Value{Type: resp.BulkString, String: "godis"}); !reflect.DeepEqual(got, want) {
		t.Errorf("GET = %+v, want %+v", got, want)
	}
}

func TestGetMissing(t *testing.T) {
	db := database.New()

	got := Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{
		bulk("GET"), bulk("nope"),
	}})
	want := resp.Value{Type: resp.BulkString, Null: true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GET missing = %+v, want %+v", got, want)
	}
}

func TestGetWrongArgs(t *testing.T) {
	db := database.New()

	got := Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{bulk("GET")}})
	if got.Type != resp.Error {
		t.Errorf("GET 参数错误 = %+v, want Error", got)
	}
}

func TestDel(t *testing.T) {
	db := database.New()

	// 删除不存在的 key → 0
	got := Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{bulk("DEL"), bulk("nope")}})
	if want := (resp.Value{Type: resp.Integer, Integer: 0}); !reflect.DeepEqual(got, want) {
		t.Errorf("DEL missing = %+v, want %+v", got, want)
	}

	// 删除存在的 key → 1
	db.Set("key", "value")
	got = Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{bulk("DEL"), bulk("key")}})
	if want := (resp.Value{Type: resp.Integer, Integer: 1}); !reflect.DeepEqual(got, want) {
		t.Errorf("DEL existing = %+v, want %+v", got, want)
	}

	// 删除后再 GET → null
	got = Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{bulk("GET"), bulk("key")}})
	want := resp.Value{Type: resp.BulkString, Null: true}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GET after DEL = %+v, want %+v", got, want)
	}
}

func TestDelWrongArgs(t *testing.T) {
	db := database.New()

	got := Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{bulk("DEL")}})
	if got.Type != resp.Error {
		t.Errorf("DEL 参数错误 = %+v, want Error", got)
	}
}

func TestCommandNameCaseInsensitive(t *testing.T) {
	db := database.New()
	db.Set("k", "v")

	for _, name := range []string{"GET", "get", "Get", "gEt"} {
		got := Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{bulk(name), bulk("k")}})
		if want := (resp.Value{Type: resp.BulkString, String: "v"}); !reflect.DeepEqual(got, want) {
			t.Errorf("命令名 %q = %+v, want %+v", name, got, want)
		}
	}
}

func TestExists(t *testing.T) {
	db := database.New()
	db.Set("a", "1")

	// 存在的 key → 1
	got := Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{bulk("EXISTS"), bulk("a")}})
	if want := (resp.Value{Type: resp.Integer, Integer: 1}); !reflect.DeepEqual(got, want) {
		t.Errorf("EXISTS a = %+v, want %+v", got, want)
	}

	// 不存在的 key → 0
	got = Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{bulk("EXISTS"), bulk("nope")}})
	if want := (resp.Value{Type: resp.Integer, Integer: 0}); !reflect.DeepEqual(got, want) {
		t.Errorf("EXISTS nope = %+v, want %+v", got, want)
	}

	// 多个 key，只数存在的 → 2
	db.Set("b", "2")
	got = Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{
		bulk("EXISTS"), bulk("a"), bulk("b"), bulk("nope"),
	}})
	if want := (resp.Value{Type: resp.Integer, Integer: 2}); !reflect.DeepEqual(got, want) {
		t.Errorf("EXISTS multi = %+v, want %+v", got, want)
	}

	// 无参数 → Error
	got = Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{bulk("EXISTS")}})
	if got.Type != resp.Error {
		t.Errorf("EXISTS 无参数 = %+v, want Error", got)
	}
}

func TestPing(t *testing.T) {
	db := database.New()

	// 无参数 → +PONG
	got := Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{bulk("PING")}})
	if want := (resp.Value{Type: resp.SimpleString, String: "PONG"}); !reflect.DeepEqual(got, want) {
		t.Errorf("PING = %+v, want %+v", got, want)
	}

	// 带参数 → bulk
	got = Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{bulk("PING"), bulk("hello")}})
	if want := (resp.Value{Type: resp.BulkString, String: "hello"}); !reflect.DeepEqual(got, want) {
		t.Errorf("PING with arg = %+v, want %+v", got, want)
	}

	// 参数多于 1 个 → Error
	got = Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{bulk("PING"), bulk("a"), bulk("b")}})
	if got.Type != resp.Error {
		t.Errorf("PING 多参数 = %+v, want Error", got)
	}
}

func TestEcho(t *testing.T) {
	db := database.New()

	got := Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{bulk("ECHO"), bulk("hello")}})
	if want := (resp.Value{Type: resp.BulkString, String: "hello"}); !reflect.DeepEqual(got, want) {
		t.Errorf("ECHO = %+v, want %+v", got, want)
	}

	// 缺参数 → Error
	got = Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{bulk("ECHO")}})
	if got.Type != resp.Error {
		t.Errorf("ECHO 缺参数 = %+v, want Error", got)
	}

	// 多参数 → Error
	got = Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{bulk("ECHO"), bulk("a"), bulk("b")}})
	if got.Type != resp.Error {
		t.Errorf("ECHO 多参数 = %+v, want Error", got)
	}
}

func TestUnknownCommand(t *testing.T) {
	db := database.New()

	got := Execute(db, resp.Value{Type: resp.Array, Array: []resp.Value{bulk("FOO")}})
	if got.Type != resp.Error {
		t.Errorf("未知命令 = %+v, want Error", got)
	}
	if got.String == "" {
		t.Error("未知命令错误消息不应为空")
	}
}

func TestInvalidRequest(t *testing.T) {
	db := database.New()

	// 非数组
	if got := Execute(db, resp.Value{Type: resp.BulkString, String: "PING"}); got.Type != resp.Error {
		t.Errorf("非数组请求 = %+v, want Error", got)
	}

	// 空数组
	if got := Execute(db, resp.Value{Type: resp.Array}); got.Type != resp.Error {
		t.Errorf("空数组请求 = %+v, want Error", got)
	}
}
