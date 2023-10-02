package redis_mutex

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RWMutex RWMutexer defines an interface for a distributed reader/writer mutex using Redis.
type RWMutex interface {
	// RLock acquires a read lock.
	RLock(ctx context.Context) error

	// RUnlock releases a read lock.
	RUnlock(ctx context.Context) error

	// Lock acquires a write lock.
	Lock(ctx context.Context) error

	// Unlock releases a write lock.
	Unlock(ctx context.Context) error
}

type KeyDBRWMutex struct {
	client    *redis.Client
	writeKey  string
	readKey   string
	writeVal  string
	readCount string
}

func NewRWMutex(client *redis.Client, baseKey string, writeVal string) *KeyDBRWMutex {
	return &KeyDBRWMutex{
		client:    client,
		writeKey:  baseKey + ":write",
		readKey:   baseKey + ":read",
		writeVal:  writeVal,
		readCount: baseKey + ":readcount",
	}
}

func (m *KeyDBRWMutex) RLock(ctx context.Context) error {
	for {
		if locked, _ := m.isWriteLocked(ctx); locked {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		// Increment read count
		m.client.Incr(ctx, m.readCount)
		break
	}
	return nil
}

func (m *KeyDBRWMutex) RUnlock(ctx context.Context) error {
	m.client.Decr(ctx, m.readCount)
	return nil
}

func (m *KeyDBRWMutex) Lock(ctx context.Context) error {
	for {
		set, err := m.client.SetNX(ctx, m.writeKey, m.writeVal, expiration).Result()
		if err != nil {
			return err
		}
		if set {
			for {
				// Wait for readers to finish
				count, _ := m.client.Get(ctx, m.readCount).Int()
				if count <= 0 {
					return nil
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (m *KeyDBRWMutex) Unlock(ctx context.Context) error {
	val, err := m.client.Get(ctx, m.writeKey).Result()
	if err != nil {
		return err
	}
	if val != m.writeVal {
		return ErrMutexOwnershipConflict
	}
	_, err = m.client.Del(ctx, m.writeKey).Result()

	return err
}

func (m *KeyDBRWMutex) isWriteLocked(ctx context.Context) (bool, error) {
	val, err := m.client.Get(ctx, m.writeKey).Result()
	if err == redis.Nil {
		return false, nil
	}

	return val == m.writeVal, err
}
