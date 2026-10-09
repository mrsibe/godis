package database

// ObjectType 标识数据库对象的数据类型。
type ObjectType uint8

const (
	StringType ObjectType = iota
	ListType
	HashType
	SetType
	ZSetType
)

// Object 保存带类型标记的值，Value 的实际类型须与 Type 一致。
type Object struct {
	Type  ObjectType
	Value any
}
