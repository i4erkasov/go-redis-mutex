package redis_mutex

import (
	"crypto/rand"
	"errors"
	"io"
	"time"
)

// expiration defines the amount of time before a mutex lock in Redis expires.
const expiration = 30000 * time.Millisecond

var (
	// ErrMutexOwnershipConflict is an error returned when trying to unlock
	// a mutex that has been released or is held by another process.
	ErrMutexOwnershipConflict = errors.New("mutex already released or acquired by someone else")
)

// uRandom generates a unique random string that can be used as a value
// for our mutex. This ensures that only the owner of the mutex can unlock it.
func uRandom() string {
	buf := make([]byte, 20)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return ""
	}

	return string(buf)
}
