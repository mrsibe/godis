package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// Writer 将 RESP2 值写入带缓冲的输出流。
type Writer struct {
	writer *bufio.Writer
}

// NewWriter 创建向 w 写入数据的序列化器。
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: bufio.NewWriter(w),
	}
}

// Write 将 value 编码为 RESP2，写入后需调用 [Writer.Flush] 刷新缓冲。
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

// Flush 将缓冲中的数据写入底层输出流。
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
