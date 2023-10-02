package redis_mutex

import (
	"context"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestKeyDBMutex_Lock(t *testing.T) {
	db, mock := redismock.NewClientMock()
	m := NewMutex(db, "testMutex")

	mock.ExpectSetNX(m.key, m.value, expiration).SetVal(true)

	err := m.Lock(context.Background())
	assert.NoError(t, err)
}

func TestKeyDBMutex_TryLock(t *testing.T) {
	db, mock := redismock.NewClientMock()
	m := NewMutex(db, "testMutex")

	mock.ExpectSetNX(m.key, m.value, expiration).SetVal(true)

	ok, err := m.TryLock(context.Background())
	assert.True(t, ok)
	assert.NoError(t, err)
}

func TestKeyDBMutex_Unlock(t *testing.T) {
	db, mock := redismock.NewClientMock()
	m := NewMutex(db, "testMutex")

	mock.ExpectGet(m.key).SetVal(m.value)
	mock.ExpectDel(m.key).SetVal(1)

	err := m.Unlock(context.Background())
	assert.NoError(t, err)
}

func TestKeyDBMutex_UnlockFail(t *testing.T) {
	db, mock := redismock.NewClientMock()
	m := NewMutex(db, "testMutex")

	mock.ExpectGet(m.key).SetVal("otherValue")

	err := m.Unlock(context.Background())
	assert.Equal(t, ErrMutexOwnershipConflict, err)
}
