package server

import (
	"io"
	"net"
	"testing"

	"github.com/mrsibe/godis/internal/database"
)

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

func TestHandleConnection(t *testing.T) {
	client, serverConn := net.Pipe()
	defer client.Close()

	s := &Server{db: database.New()}
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.handleConnection(serverConn)
	}()

	// PING
	if _, err := client.Write([]byte("*1\r\n$4\r\nPING\r\n")); err != nil {
		t.Fatal(err)
	}
	expect(t, client, "+PONG\r\n")

	// ECHO hi
	if _, err := client.Write([]byte("*2\r\n$4\r\nECHO\r\n$2\r\nhi\r\n")); err != nil {
		t.Fatal(err)
	}
	expect(t, client, "$2\r\nhi\r\n")

	// SET foo bar
	if _, err := client.Write([]byte("*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n")); err != nil {
		t.Fatal(err)
	}
	expect(t, client, "+OK\r\n")

	// GET foo
	if _, err := client.Write([]byte("*2\r\n$3\r\nGET\r\n$3\r\nfoo\r\n")); err != nil {
		t.Fatal(err)
	}
	expect(t, client, "$3\r\nbar\r\n")

	// GET nope
	if _, err := client.Write([]byte("*2\r\n$3\r\nGET\r\n$4\r\nnope\r\n")); err != nil {
		t.Fatal(err)
	}
	expect(t, client, "$-1\r\n")

	// DEL foo
	if _, err := client.Write([]byte("*2\r\n$3\r\nDEL\r\n$3\r\nfoo\r\n")); err != nil {
		t.Fatal(err)
	}
	expect(t, client, ":1\r\n")

	// 再次 DEL foo
	if _, err := client.Write([]byte("*2\r\n$3\r\nDEL\r\n$3\r\nfoo\r\n")); err != nil {
		t.Fatal(err)
	}
	expect(t, client, ":0\r\n")

	// 未知命令 FOO
	if _, err := client.Write([]byte("*1\r\n$3\r\nFOO\r\n")); err != nil {
		t.Fatal(err)
	}
	expect(t, client, "-ERR unknown command 'FOO'\r\n")

	client.Close()
	<-done
}

func TestHandleConnectionCloseMidCommand(t *testing.T) {
	client, serverConn := net.Pipe()

	s := &Server{db: database.New()}
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.handleConnection(serverConn)
	}()

	// 命令未发送完整即断开连接。
	if _, err := client.Write([]byte("*2\r\n$3\r\nGET")); err != nil {
		t.Fatal(err)
	}
	client.Close()

	<-done
}
