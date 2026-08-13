package server

import (
	"io"
	"net"
	"testing"

	"github.com/mrsibe/godis/internal/database"
)

// expect 从 r 中精确读取 want 长度的字节并比对。
func expect(t *testing.T, r io.Reader, want string) {
	t.Helper()

	buf := make([]byte, len(want))
	if _, err := io.ReadFull(r, buf); err != nil {
		t.Fatalf("ReadFull: %v", err)
	}
	if string(buf) != want {
		t.Errorf("got %q, want %q", string(buf), want)
	}
}

// TestHandleConnection 用 net.Pipe 模拟客户端，走完整的 RESP 往返。
func TestHandleConnection(t *testing.T) {
	client, serverConn := net.Pipe()
	defer client.Close()

	s := &Server{db: database.New()}
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.handleConnection(serverConn)
	}()

	// PING → +PONG
	if _, err := client.Write([]byte("*1\r\n$4\r\nPING\r\n")); err != nil {
		t.Fatal(err)
	}
	expect(t, client, "+PONG\r\n")

	// ECHO hi → $2\r\nhi\r\n
	if _, err := client.Write([]byte("*2\r\n$4\r\nECHO\r\n$2\r\nhi\r\n")); err != nil {
		t.Fatal(err)
	}
	expect(t, client, "$2\r\nhi\r\n")

	// SET foo bar → +OK
	if _, err := client.Write([]byte("*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")); err != nil {
		t.Fatal(err)
	}
	expect(t, client, "+OK\r\n")

	// GET foo → $3\r\nbar\r\n
	if _, err := client.Write([]byte("*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n")); err != nil {
		t.Fatal(err)
	}
	expect(t, client, "$3\r\nbar\r\n")

	// GET 不存在的 key → $-1\r\n
	if _, err := client.Write([]byte("*2\r\n$3\r\nGET\r\n$4\r\nnope\r\n")); err != nil {
		t.Fatal(err)
	}
	expect(t, client, "$-1\r\n")

	// DEL foo → :1\r\n
	if _, err := client.Write([]byte("*2\r\n$3\r\nDEL\r\n$3\r\nfoo\r\n")); err != nil {
		t.Fatal(err)
	}
	expect(t, client, ":1\r\n")

	// 再 DEL foo → :0\r\n
	if _, err := client.Write([]byte("*2\r\n$3\r\nDEL\r\n$3\r\nfoo\r\n")); err != nil {
		t.Fatal(err)
	}
	expect(t, client, ":0\r\n")

	// 未知命令 → -ERR unknown command 'FOO'\r\n
	if _, err := client.Write([]byte("*1\r\n$3\r\nFOO\r\n")); err != nil {
		t.Fatal(err)
	}
	expect(t, client, "-ERR unknown command 'FOO'\r\n")

	// 关闭连接，server 应优雅退出
	client.Close()
	<-done
}

// TestHandleConnectionCloseMidCommand 验证半截命令 + 断开时 server 不 panic、正常退出。
func TestHandleConnectionCloseMidCommand(t *testing.T) {
	client, serverConn := net.Pipe()

	s := &Server{db: database.New()}
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.handleConnection(serverConn)
	}()

	// 发一半命令就关闭连接
	if _, err := client.Write([]byte("*2\r\n$3\r\nGET")); err != nil {
		t.Fatal(err)
	}
	client.Close()

	// server 应检测到断开并退出，而不是阻塞或 panic
	<-done
}
