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

// Server 通过 TCP 处理 RESP 命令，所有连接共享同一数据库。
type Server struct {
	addr string
	db   *database.DB
}

// New 创建监听地址为 addr、数据库为空的服务端。
func New(addr string) *Server {
	return &Server{
		addr: addr,
		db:   database.New(),
	}
}

// Start 启动监听并为每个连接创建独立的 goroutine。
// 监听失败时返回错误，接受连接失败时记录日志并继续。
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
			// 正常断开连接不记录错误。
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
