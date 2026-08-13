# godis

A Redis-compatible in-memory key-value store written in Go.

## Features

- RESP protocol parsing and serialization
- Concurrency-safe storage
- Pluggable command dispatch (add a command in one line)

Supported commands: `SET`, `GET`, `DEL`, `EXISTS`, `PING`, `ECHO`

## Get Started

```bash
go run ./cmd/server
```

The server listens on `:6379` by default.
