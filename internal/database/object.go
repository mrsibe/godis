package database

type ObjectType uint8

const (
	StringType ObjectType = iota
	ListType
	HashType
	SetType
	ZSetType
)

type Object struct {
	Type  ObjectType
	Value any
}
