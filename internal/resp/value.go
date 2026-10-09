package resp

type Type byte

const (
	SimpleString Type = '+'
	Error Type = '-'
	Integer Type = ':'
	BulkString Type = '$'
	Array Type = '*'
)

type Value struct {
	Type Type

	String  string
	Integer int
	Array   []Value

	Null bool
}
