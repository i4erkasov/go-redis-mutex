# Go Redis Mutex

![GitHub issues](https://img.shields.io/github/issues/i4erkasov/go-redis-mutex)
![GitHub stars](https://img.shields.io/github/stars/i4erkasov/go-redis-mutex)
![GitHub forks](https://img.shields.io/github/forks/i4erkasov/go-redis-mutex)
![GitHub last commit](https://img.shields.io/github/last-commit/i4erkasov/go-redis-mutex)
![GoDoc](https://godoc.org/github.com/i4erkasov/go-redis-mutex?status.svg)


`go-redis-mutex` is a Go package that offers an implementation of a distributed mutex based on Redis. This package provides both a standard mutex and a read/write mutex.

Locking implementation recommendations were taken from [KeyDB documentation](https://docs.keydb.dev/docs/distlock/).

## Dependencies

- [go-redis](https://github.com/go-redis/redis) (version 9)

## Installation

```bash
go get github.com/i4erkasov/go-redis-mutex
```

## Usage

```go
import "github.com/i4erkasov/go-redis-mutex"

// Create a new mutex
mutex := redis_mutex.NewMutex(client, "myMutexKey")

// Attempt to lock
err := mutex.Lock(context.Background())

// Attempt to unlock
err = mutex.Unlock(context.Background())
```