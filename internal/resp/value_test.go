package resp

import "testing"

// TestTypeConstants 验证各 Type 常量对应正确的 RESP2 协议字节。
func TestTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		got  Type
		want byte
	}{
		{name: "SimpleString", got: SimpleString, want: '+'},
		{name: "Error", got: Error, want: '-'},
		{name: "Integer", got: Integer, want: ':'},
		{name: "BulkString", got: BulkString, want: '$'},
		{name: "Array", got: Array, want: '*'},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if byte(tt.got) != tt.want {
				t.Errorf("Type = %q (%d), want %q (%d)", tt.got, tt.got, tt.want, tt.want)
			}
		})
	}
}

// TestValueZeroValue 验证 Value 的零值是合理的中性状态。
func TestValueZeroValue(t *testing.T) {
	var v Value

	if v.Type != 0 {
		t.Errorf("Type = %q (%d), want zero value", v.Type, v.Type)
	}
	if v.String != "" {
		t.Errorf("String = %q, want empty string", v.String)
	}
	if v.Integer != 0 {
		t.Errorf("Integer = %d, want 0", v.Integer)
	}
	if v.Array != nil {
		t.Errorf("Array = %v, want nil", v.Array)
	}
	if v.Null {
		t.Errorf("Null = true, want false")
	}
}

// TestValueFields 验证各种 RESP 类型的 Value 构造后字段正确。
func TestValueFields(t *testing.T) {
	tests := []struct {
		name  string
		value Value
	}{
		{name: "SimpleString", value: Value{Type: SimpleString, String: "OK"}},
		{name: "Error", value: Value{Type: Error, String: "ERR unknown command"}},
		{name: "Integer", value: Value{Type: Integer, Integer: 42}},
		{name: "IntegerNegative", value: Value{Type: Integer, Integer: -1}},
		{name: "BulkString", value: Value{Type: BulkString, String: "value"}},
		{name: "BulkStringEmpty", value: Value{Type: BulkString, String: ""}},
		{name: "ArrayEmpty", value: Value{Type: Array}},
		{name: "ArrayNested", value: Value{
			Type: Array,
			Array: []Value{
				{Type: SimpleString, String: "a"},
				{Type: Integer, Integer: 1},
				{Type: Array, Array: []Value{{Type: BulkString, String: "b"}}},
			},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 此处没有序列化方法，仅验证构造的 Value 字段语义
			switch tt.value.Type {
			case Integer:
				if tt.value.String != "" {
					t.Errorf("Integer Value 不应带 String 字段, got %q", tt.value.String)
				}
				_ = tt.value.Integer
			case Array:
				if tt.value.String != "" || tt.value.Integer != 0 {
					t.Errorf("Array Value 不应带 String/Integer 字段, got %q/%d", tt.value.String, tt.value.Integer)
				}
				_ = tt.value.Array
			default:
				// 字符串类类型只应使用 String 字段
				if tt.value.Integer != 0 || len(tt.value.Array) != 0 {
					t.Errorf("%q Value 不应带 Integer/Array 字段, got %d/%v", tt.value.Type, tt.value.Integer, tt.value.Array)
				}
			}
		})
	}
}

// TestValueArrayNested 验证嵌套数组的深度结构。
func TestValueArrayNested(t *testing.T) {
	value := Value{
		Type: Array,
		Array: []Value{
			{Type: SimpleString, String: "PING"},
			{Type: Array, Array: []Value{{Type: BulkString, String: "inner"}}},
		},
	}

	if value.Type != Array {
		t.Fatalf("外层 Type = %q, want Array", value.Type)
	}
	if len(value.Array) != 2 {
		t.Fatalf("外层 Array 长度 = %d, want 2", len(value.Array))
	}

	first := value.Array[0]
	if first.Type != SimpleString || first.String != "PING" {
		t.Errorf("Array[0] = {%q %q}, want SimpleString \"PING\"", first.Type, first.String)
	}

	inner := value.Array[1]
	if inner.Type != Array || len(inner.Array) != 1 {
		t.Fatalf("Array[1] 应为含 1 个元素的数组, got Type=%q len=%d", inner.Type, len(inner.Array))
	}
	leaf := inner.Array[0]
	if leaf.Type != BulkString || leaf.String != "inner" {
		t.Errorf("叶子元素 = {%q %q}, want BulkString \"inner\"", leaf.Type, leaf.String)
	}
}
