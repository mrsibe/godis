package resp

// Type 是 RESP 协议的类型标记，值对应协议中的类型前缀字节。
type Type byte

const (
	// SimpleString 表示以 + 开头的简单字符串。
	SimpleString Type = '+'
	// Error 表示以 - 开头的错误消息。
	Error Type = '-'
	// Integer 表示以 : 开头的整数。
	Integer Type = ':'
	// BulkString 表示以 $ 开头的批量字符串。
	BulkString Type = '$'
	// Array 表示以 * 开头的数组。
	Array Type = '*'
)

// Value 表示一个 RESP 值，具体使用哪个字段由 Type 决定：
//   - SimpleString/Error/BulkString 使用 String；
//   - Integer 使用 Integer；
//   - Array 使用 Array。
//
// Null 为 true 时表示 RESP 的 null 值（序列化为 $-1 或 *-1）。
type Value struct {
	Type Type

	String  string
	Integer int
	Array   []Value

	Null bool
}
