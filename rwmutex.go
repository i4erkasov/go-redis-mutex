package redis_mutex

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RWMutex interface {
	RLock(ctx context.Context) error
	RUnlock(ctx context.Context)
	Lock(ctx context.Context) error
	Unlock(ctx context.Context) error
}

type KeyDBRWMutex struct {
	client    *redis.Client
	writeKey  string
	readKey   string
	writeVal  string
	readCount string
}

// NewRWMutex initializes a new Redis-based reader-writer mutex.
// The mutex separates the keys for read and write locks for granular control.
// It uses a specific value for write locks to ensure that only the owner can unlock it,
// while read locks increment a counter to manage multiple readers.
func NewRWMutex(client *redis.Client, baseKey string, writeVal string) *KeyDBRWMutex {
	return &KeyDBRWMutex{
		client:    client,
		writeKey:  baseKey + ":write",
		readKey:   baseKey + ":read",
		writeVal:  writeVal,
		readCount: baseKey + ":readcount",
	}
}

// RLock attempts to acquire a read lock, waiting if a write lock is held.
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

// RUnlock releases a read lock by decrementing the read count.
func (m *KeyDBRWMutex) RUnlock(ctx context.Context) {
	m.client.Decr(ctx, m.readCount)
}

// Lock attempts to acquire a write lock, waiting if there are readers or another writer.
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

// Unlock releases a write lock, but only if the caller owns the lock.
func (m *KeyDBRWMutex) Unlock(ctx context.Context) error {
	val, err := m.client.Get(ctx, m.writeKey).Result()
	// If the key doesn't exist, there's nothing to unlock
	if err == redis.Nil {
		return nil
	}

	if err != nil {
		return err
	}

	if val != m.writeVal {
		return ErrMutexOwnershipConflict
	}

	return m.client.Del(ctx, m.writeKey).Err()
}

// isWriteLocked checks if a write lock is currently held.
func (m *KeyDBRWMutex) isWriteLocked(ctx context.Context) (bool, error) {
	val, err := m.client.Get(ctx, m.writeKey).Result()
	if err == redis.Nil {
		return false, nil
	}

	return val == m.writeVal, err
}
