package redis_mutex

import (
	"context"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestKeyDBRWMutex_RLock(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mutex := NewRWMutex(db, "testRWMutex", "writeVal")

	mock.ExpectGet(mutex.writeKey).SetVal("")
	mock.ExpectIncr(mutex.readCount)

	err := mutex.RLock(context.Background())
	assert.NoError(t, err)
}

func TestKeyDBRWMutex_RUnlock(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mutex := NewRWMutex(db, "testRWMutex", "writeVal")

	mock.ExpectDecr(mutex.readCount)

	err := mutex.RUnlock(context.Background())
	assert.NoError(t, err)
}

func TestKeyDBRWMutex_Lock(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mutex := NewRWMutex(db, "testRWMutex", "writeVal")

	mock.ExpectSetNX(mutex.writeKey, mutex.writeVal, expiration).SetVal(true)
	mock.ExpectGet(mutex.readCount).SetVal("0")

	err := mutex.Lock(context.Background())
	assert.NoError(t, err)
}

func TestKeyDBRWMutex_Unlock(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mutex := NewRWMutex(db, "testRWMutex", "writeVal")

	mock.ExpectGet(mutex.writeKey).SetVal("writeVal")
	mock.ExpectDel(mutex.writeKey)

	err := mutex.Unlock(context.Background())
	assert.NoError(t, err)
}

func TestKeyDBRWMutex_UnlockFail(t *testing.T) {
	db, mock := redismock.NewClientMock()
	mutex := NewRWMutex(db, "testRWMutex", "writeVal")

	mock.ExpectGet(mutex.writeKey).SetVal("otherVal")

	err := mutex.Unlock(context.Background())
	assert.Equal(t, ErrMutexOwnershipConflict, err)
}
