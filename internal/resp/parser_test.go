package resp

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseBulkString(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Value
	}{
		{name: "simple", input: "$5\r\nhello\r\n", want: Value{Type: BulkString, String: "hello"}},
		{name: "empty", input: "$0\r\n\r\n", want: Value{Type: BulkString, String: ""}},
		{name: "null", input: "$-1\r\n", want: Value{Type: BulkString, Null: true}},
		// bulk string 内部可以包含 \r\n，不受 readLine 的换行切割影响
		{name: "withCRLF", input: "$4\r\na\r\nb\r\n", want: Value{Type: BulkString, String: "a\r\nb"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(strings.NewReader(tt.input))
			got, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseArray(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Value
	}{
		{
			name:  "twoBulks",
			input: "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
			want: Value{Type: Array, Array: []Value{
				{Type: BulkString, String: "foo"},
				{Type: BulkString, String: "bar"},
			}},
		},
		{
			name:  "empty",
			input: "*0\r\n",
			want:  Value{Type: Array, Array: []Value{}},
		},
		{
			name:  "null",
			input: "*-1\r\n",
			want:  Value{Type: Array, Null: true},
		},
		{
			name:  "nested",
			input: "*1\r\n*2\r\n$1\r\na\r\n$1\r\nb\r\n",
			want: Value{Type: Array, Array: []Value{
				{Type: Array, Array: []Value{
					{Type: BulkString, String: "a"},
					{Type: BulkString, String: "b"},
				}},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(strings.NewReader(tt.input))
			got, err := p.Parse()
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseUnsupportedType(t *testing.T) {
	// SimpleString、Error、Integer 的类型常量已定义，但 parser 尚未实现它们
	inputs := []string{
		"+OK\r\n",
		"-ERR unknown command\r\n",
		":42\r\n",
	}

	for _, input := range inputs {
		t.Run(input[:1], func(t *testing.T) {
			p := NewParser(strings.NewReader(input))
			_, err := p.Parse()
			if err == nil {
				t.Errorf("Parse(%q) error = nil, want error", input)
			}
		})
	}
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{name: "empty", input: ""},
		{name: "truncatedBulk", input: "$5\r\nhel"},
		{name: "truncatedArray", input: "*2\r\n$3\r\nfoo"},
		{name: "invalidLength", input: "*abc\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(strings.NewReader(tt.input))
			_, err := p.Parse()
			if err == nil {
				t.Errorf("Parse(%q) error = nil, want error", tt.input)
			}
		})
	}
}

func TestParseMultipleCommands(t *testing.T) {
	// 模拟一个 TCP 连接里连续到达多条命令
	p := NewParser(strings.NewReader("$3\r\nfoo\r\n$3\r\nbar\r\n"))

	first, err := p.Parse()
	if err != nil {
		t.Fatalf("first Parse() error = %v", err)
	}
	if want := (Value{Type: BulkString, String: "foo"}); !reflect.DeepEqual(first, want) {
		t.Errorf("first Parse() = %+v, want %+v", first, want)
	}

	second, err := p.Parse()
	if err != nil {
		t.Fatalf("second Parse() error = %v", err)
	}
	if want := (Value{Type: BulkString, String: "bar"}); !reflect.DeepEqual(second, want) {
		t.Errorf("second Parse() = %+v, want %+v", second, want)
	}
}
