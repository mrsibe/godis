package resp

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Parser 是 RESP 协议的流式解析器，按类型前缀递归解析输入流。
type Parser struct {
	reader *bufio.Reader
}

// NewParser 基于 rd 创建一个 Parser。
func NewParser(rd io.Reader) *Parser {
	return &Parser{
		reader: bufio.NewReader(rd),
	}
}

// Parse 读取并解析一个完整的 RESP 值。
func (p *Parser) Parse() (Value, error) {
	prefix, err := p.reader.ReadByte()
	if err != nil {
		return Value{}, err
	}

	switch Type(prefix) {
	case Array:
		return p.parseArray()
	case BulkString:
		return p.parseBulkString()
	default:
		return Value{}, fmt.Errorf("unsupported RESP type: %q", prefix)
	}
}

func (p *Parser) parseArray() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}

	length, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, err
	}

	// 负长度（*-1）表示 null array
	if length < 0 {
		return Value{Type: Array, Null: true}, nil
	}

	values := make([]Value, 0, length)

	for range length {
		value, err := p.Parse()
		if err != nil {
			return Value{}, err
		}

		values = append(values, value)
	}

	return Value{
		Type:  Array,
		Array: values,
	}, nil
}

func (p *Parser) parseBulkString() (Value, error) {
	line, err := p.readLine()
	if err != nil {
		return Value{}, err
	}

	length, err := strconv.Atoi(line)
	if err != nil {
		return Value{}, err
	}

	// 负长度（$-1）表示 null bulk string
	if length < 0 {
		return Value{
			Type: BulkString,
			Null: true,
		}, nil
	}

	buf := make([]byte, length+2)

	if _, err := io.ReadFull(p.reader, buf); err != nil {
		return Value{}, err
	}

	return Value{
		Type:   BulkString,
		String: string(buf[:length]),
	}, nil
}

func (p *Parser) readLine() (string, error) {
	line, err := p.reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	line = strings.TrimSuffix(line, "\r\n")

	return line, nil
}
