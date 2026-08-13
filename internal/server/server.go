package server

import (
	"errors"
	"io"
	"log"
	"net"

	"github.com/mrsibe/godis/internal/command"
	"github.com/mrsibe/godis/internal/database"
	"github.com/mrsibe/godis/internal/resp"
)

// Server 是 godis 的 TCP 服务端。
type Server struct {
	addr string
	db   *database.DB
}

// New 基于监听地址 addr 创建一个 Server，并初始化空数据库。
func New(addr string) *Server {
	return &Server{
		addr: addr,
		db:   database.New(),
	}
}

// Start 监听 addr 并循环处理连接，直到返回错误。
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("godis listening on %s", s.addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept connection error: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	parser := resp.NewParser(conn)
	writer := resp.NewWriter(conn)

	for {
		value, err := parser.Parse()
		if err != nil {
			// 客户端正常断开连接时静默退出
			if errors.Is(err, io.EOF) {
				return
			}
			log.Printf("parse error: %v", err)
			return
		}

		reply := command.Execute(s.db, value)
		if err := writer.Write(reply); err != nil {
			log.Printf("write error: %v", err)
			return
		}
		if err := writer.Flush(); err != nil {
			log.Printf("flush error: %v", err)
			return
		}
	}
}
