package command

import (
	"strings"

	"github.com/mrsibe/godis/internal/database"
	"github.com/mrsibe/godis/internal/resp"
)

// Handler 是单个命令的执行函数，接收数据库和参数，返回 RESP 回复。
type Handler func(db *database.DB, args []resp.Value) resp.Value

// handlers 注册所有命令，命令名统一为大写。
// 新增命令时只需实现一个 Handler 并在这里登记一行。
var handlers = map[string]Handler{
	"PING":   ping,
	"ECHO":   echo,
	"SET":    set,
	"GET":    get,
	"DEL":    del,
	"EXISTS": exists,
}

// Execute 分发并执行一条命令，返回 RESP 回复。
// req 是由 Parser 解析出的 Array，首元素为命令名，其余为参数。
func Execute(db *database.DB, req resp.Value) resp.Value {
	if req.Type != resp.Array || len(req.Array) == 0 {
		return resp.Value{Type: resp.Error, String: "ERR invalid command"}
	}

	handler, ok := handlers[strings.ToUpper(req.Array[0].String)]
	if !ok {
		return resp.Value{Type: resp.Error, String: "ERR unknown command '" + req.Array[0].String + "'"}
	}

	return handler(db, req.Array[1:])
}

func ping(_ *database.DB, args []resp.Value) resp.Value {
	if len(args) > 1 {
		return resp.Value{Type: resp.Error, String: "ERR wrong number of arguments for 'ping' command"}
	}

	if len(args) == 1 {
		return resp.Value{Type: resp.BulkString, String: args[0].String}
	}

	return resp.Value{Type: resp.SimpleString, String: "PONG"}
}

func echo(_ *database.DB, args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Type: resp.Error, String: "ERR wrong number of arguments for 'echo' command"}
	}

	return resp.Value{Type: resp.BulkString, String: args[0].String}
}

func set(db *database.DB, args []resp.Value) resp.Value {
	if len(args) != 2 {
		return resp.Value{Type: resp.Error, String: "ERR wrong number of arguments for 'set' command"}
	}

	db.Set(args[0].String, &database.Object{
		Type:  database.StringType,
		Value: args[1].String,
	})

	return resp.Value{Type: resp.SimpleString, String: "OK"}
}

func get(db *database.DB, args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Type: resp.Error, String: "ERR wrong number of arguments for 'get' command"}
	}

	obj, ok := db.Get(args[0].String)
	if !ok {
		return resp.Value{Type: resp.BulkString, Null: true}
	}
	if obj.Type != database.StringType {
		return resp.Value{Type: resp.Error,
			String: "WRONGTYPE Operation against a key holding the wrong kind of value"}
	}

	return resp.Value{Type: resp.BulkString, String: obj.Value.(string)}
}

func del(db *database.DB, args []resp.Value) resp.Value {
	if len(args) != 1 {
		return resp.Value{Type: resp.Error, String: "ERR wrong number of arguments for 'del' command"}
	}

	if db.Del(args[0].String) {
		return resp.Value{Type: resp.Integer, Integer: 1}
	}

	return resp.Value{Type: resp.Integer, Integer: 0}
}

func exists(db *database.DB, args []resp.Value) resp.Value {
	if len(args) == 0 {
		return resp.Value{Type: resp.Error, String: "ERR wrong number of arguments for 'exists' command"}
	}

	count := 0
	for _, arg := range args {
		if _, ok := db.Get(arg.String); ok {
			count++
		}
	}

	return resp.Value{Type: resp.Integer, Integer: count}
}
