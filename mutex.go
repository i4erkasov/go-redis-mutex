package redis_mutex

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Mutex interface {
	Lock(ctx context.Context) error
	Unlock(ctx context.Context) error
	TryLock(ctx context.Context) (bool, error)
}

// KeyDBMutex provides a mutex mechanism using Redis.
type KeyDBMutex struct {
	client *redis.Client // The Redis client
	key    string        // The key used to identify the mutex in Redis
	value  string        // The value that represents the owner of the mutex
}

// NewMutex initializes a new Redis-based mutex.
// The mutex uses a random value to ensure that only the owner can unlock it.
func NewMutex(client *redis.Client, key string) *KeyDBMutex {
	return &KeyDBMutex{
		client: client,
		key:    key,
		value:  uRandom(),
	}
}

// Lock tries to obtain the lock. If it's already held, it will keep trying
// at intervals until the lock is acquired.
func (m *KeyDBMutex) Lock(ctx context.Context) error {
	for {
		set, err := m.client.SetNX(ctx, m.key, m.value, expiration).Result()
		if err != nil {
			return err
		}
		if set {
			return nil
		}
		// Sleep for a short duration before retrying
		time.Sleep(100 * time.Millisecond)
	}
}

// TryLock attempts to acquire the lock once. If successful, it returns true.
func (m *KeyDBMutex) TryLock(ctx context.Context) (bool, error) {
	return m.client.SetNX(ctx, m.key, m.value, expiration).Result()
}

// Unlock releases the lock. If the lock is held by another value
// (i.e., acquired by another process or thread), it returns an error.
func (m *KeyDBMutex) Unlock(ctx context.Context) error {
	// Fetch the current value associated with the mutex key
	val, err := m.client.Get(ctx, m.key).Result()
	// If the key doesn't exist, there's nothing to unlock
	if err == redis.Nil {
		return nil
	}

	// For any other Redis error, return it
	if err != nil {
		return err
	}

	// Check if the current value matches our own mutex value
	if val != m.value {
		return ErrMutexOwnershipConflict
	}

	// If we own the mutex, delete the key to release the lock
	_, err = m.client.Del(ctx, m.key).Result()

	return err
}
