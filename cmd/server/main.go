package main

import (
	"log"

	"github.com/mrsibe/godis/internal/server"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	return server.New(":6379").Start()
}
