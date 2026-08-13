package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// Writer 是 RESP 协议的序列化器，将 Value 写入底层 io.Writer。
type Writer struct {
	writer *bufio.Writer
}

// NewWriter 基于 w 创建一个 Writer。
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: bufio.NewWriter(w),
	}
}

// Write 将一个 Value 序列化为 RESP2 协议写入缓冲。
// 写入完成后需要调用 Flush 才会真正落到下层 io.Writer。
func (w *Writer) Write(value Value) error {
	if value.Null {
		if value.Type == Array {
			return w.writeString("*-1")
		}
		return w.writeString("$-1")
	}

	switch value.Type {
	case SimpleString:
		return w.writeString("+" + value.String)
	case Error:
		return w.writeString("-" + value.String)
	case Integer:
		return w.writeString(":" + strconv.Itoa(value.Integer))
	case BulkString:
		return w.writeBulkString(value.String)
	case Array:
		return w.writeArray(value.Array)
	default:
		return fmt.Errorf("unsupported RESP type: %q", value.Type)
	}
}

func (w *Writer) Flush() error {
	return w.writer.Flush()
}

func (w *Writer) writeString(s string) error {
	_, err := w.writer.WriteString(s + "\r\n")
	return err
}

func (w *Writer) writeBulkString(s string) error {
	if err := w.writeString("$" + strconv.Itoa(len(s))); err != nil {
		return err
	}
	_, err := w.writer.WriteString(s + "\r\n")
	return err
}

func (w *Writer) writeArray(values []Value) error {
	if err := w.writeString("*" + strconv.Itoa(len(values))); err != nil {
		return err
	}
	for _, v := range values {
		if err := w.Write(v); err != nil {
			return err
		}
	}
	return nil
}
