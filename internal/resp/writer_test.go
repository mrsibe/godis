package resp

import (
	"bytes"
	"testing"
)

func writeToString(t *testing.T, values ...Value) string {
	t.Helper()

	var buf bytes.Buffer
	w := NewWriter(&buf)
	for _, v := range values {
		if err := w.Write(v); err != nil {
			t.Fatalf("Write() error = %v", err)
		}
	}
	if err := w.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}
	return buf.String()
}

func TestWriteSimpleString(t *testing.T) {
	got := writeToString(t, Value{Type: SimpleString, String: "OK"})
	if want := "+OK\r\n"; got != want {
		t.Errorf("Write() = %q, want %q", got, want)
	}
}

func TestWriteError(t *testing.T) {
	got := writeToString(t, Value{Type: Error, String: "ERR unknown command"})
	if want := "-ERR unknown command\r\n"; got != want {
		t.Errorf("Write() = %q, want %q", got, want)
	}
}

func TestWriteInteger(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		want  string
	}{
		{name: "positive", value: Value{Type: Integer, Integer: 42}, want: ":42\r\n"},
		{name: "negative", value: Value{Type: Integer, Integer: -1}, want: ":-1\r\n"},
		{name: "zero", value: Value{Type: Integer, Integer: 0}, want: ":0\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := writeToString(t, tt.value); got != tt.want {
				t.Errorf("Write() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteBulkString(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		want  string
	}{
		{name: "simple", value: Value{Type: BulkString, String: "hello"}, want: "$5\r\nhello\r\n"},
		{name: "empty", value: Value{Type: BulkString, String: ""}, want: "$0\r\n\r\n"},
		{name: "withCRLF", value: Value{Type: BulkString, String: "a\r\nb"}, want: "$4\r\na\r\nb\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := writeToString(t, tt.value); got != tt.want {
				t.Errorf("Write() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteArray(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		want  string
	}{
		{
			name: "twoBulks",
			value: Value{Type: Array, Array: []Value{
				{Type: BulkString, String: "foo"},
				{Type: BulkString, String: "bar"},
			}},
			want: "*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n",
		},
		{
			name:  "empty",
			value: Value{Type: Array},
			want:  "*0\r\n",
		},
		{
			name: "mixed",
			value: Value{Type: Array, Array: []Value{
				{Type: SimpleString, String: "OK"},
				{Type: Integer, Integer: 7},
			}},
			want: "*2\r\n+OK\r\n:7\r\n",
		},
		{
			name: "nested",
			value: Value{Type: Array, Array: []Value{
				{Type: Array, Array: []Value{{Type: BulkString, String: "a"}}},
			}},
			want: "*1\r\n*1\r\n$1\r\na\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := writeToString(t, tt.value); got != tt.want {
				t.Errorf("Write() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteNull(t *testing.T) {
	tests := []struct {
		name  string
		value Value
		want  string
	}{
		{name: "bulkString", value: Value{Type: BulkString, Null: true}, want: "$-1\r\n"},
		{name: "array", value: Value{Type: Array, Null: true}, want: "*-1\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := writeToString(t, tt.value); got != tt.want {
				t.Errorf("Write() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWriteMultipleValues(t *testing.T) {
	got := writeToString(t,
		Value{Type: Integer, Integer: 1},
		Value{Type: BulkString, String: "x"},
	)
	if want := ":1\r\n$1\r\nx\r\n"; got != want {
		t.Errorf("Write() = %q, want %q", got, want)
	}
}

func TestWriteUnsupportedType(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	if err := w.Write(Value{Type: Type(0)}); err == nil {
		t.Error("Write(unknown type) error = nil, want error")
	}
}
