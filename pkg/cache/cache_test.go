package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryCache_GetSet(t *testing.T) {
	c := NewInMemory()
	ctx := context.Background()

	// Set & Get
	err := c.Set(ctx, "key1", "value1", time.Minute)
	require.NoError(t, err)

	val, err := c.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)
}

func TestInMemoryCache_Miss(t *testing.T) {
	c := NewInMemory()
	ctx := context.Background()

	val, err := c.Get(ctx, "nonexistent")
	assert.ErrorIs(t, err, ErrMiss)
	assert.Empty(t, val)
}

func TestInMemoryCache_TTLExpired(t *testing.T) {
	c := NewInMemory()
	ctx := context.Background()

	// Set with 0 TTL = no expiry
	err := c.Set(ctx, "perm", "forever", 0)
	require.NoError(t, err)

	val, err := c.Get(ctx, "perm")
	assert.NoError(t, err)
	assert.Equal(t, "forever", val)

	// Set with 1 nanosecond — should be expired by the time we get
	err = c.Set(ctx, "gone", "bye", 1*time.Nanosecond)
	require.NoError(t, err)

	// tiny sleep to ensure expiry
	time.Sleep(10 * time.Millisecond)

	val, err = c.Get(ctx, "gone")
	assert.ErrorIs(t, err, ErrMiss)
	assert.Empty(t, val)
}

func TestInMemoryCache_Del(t *testing.T) {
	c := NewInMemory()
	ctx := context.Background()

	_ = c.Set(ctx, "delme", "value", time.Minute)
	_ = c.Del(ctx, "delme")

	val, err := c.Get(ctx, "delme")
	assert.ErrorIs(t, err, ErrMiss)
	assert.Empty(t, val)
}

func TestInMemoryCache_Exists(t *testing.T) {
	c := NewInMemory()
	ctx := context.Background()

	ok, err := c.Exists(ctx, "missing")
	assert.NoError(t, err)
	assert.False(t, ok)

	_ = c.Set(ctx, "present", "yep", time.Minute)

	ok, err = c.Exists(ctx, "present")
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestInMemoryCache_Close(t *testing.T) {
	c := NewInMemory()
	ctx := context.Background()

	_ = c.Set(ctx, "a", "1", time.Minute)
	err := c.Close()
	assert.NoError(t, err)

	// After close, data should be nil, but Get won't panic because of RLock
	val, err := c.Get(ctx, "a")
	assert.ErrorIs(t, err, ErrMiss)
	assert.Empty(t, val)
}

func TestNoopCache(t *testing.T) {
	c := NewNoop()
	ctx := context.Background()

	val, err := c.Get(ctx, "anything")
	assert.ErrorIs(t, err, ErrMiss)
	assert.Empty(t, val)

	err = c.Set(ctx, "k", "v", time.Minute)
	assert.NoError(t, err)

	err = c.Del(ctx, "k")
	assert.NoError(t, err)

	ok, err := c.Exists(ctx, "k")
	assert.NoError(t, err)
	assert.False(t, ok)

	err = c.Close()
	assert.NoError(t, err)
}

func TestNoopCache_ImplementsInterface(t *testing.T) {
	// Compile-time check
	var _ Cache = (*NoopCache)(nil)
	var _ Cache = (*InMemoryCache)(nil)
}

func TestInMemoryCache_MultipleKeys(t *testing.T) {
	c := NewInMemory()
	ctx := context.Background()

	_ = c.Set(ctx, "a", "1", time.Minute)
	_ = c.Set(ctx, "b", "2", time.Minute)
	_ = c.Set(ctx, "c", "3", time.Minute)

	// Delete multiple
	_ = c.Del(ctx, "a", "c")

	_, err := c.Get(ctx, "a")
	assert.ErrorIs(t, err, ErrMiss)

	val, err := c.Get(ctx, "b")
	assert.NoError(t, err)
	assert.Equal(t, "2", val)

	_, err = c.Get(ctx, "c")
	assert.ErrorIs(t, err, ErrMiss)
}

func TestInMemoryCache_Overwrite(t *testing.T) {
	c := NewInMemory()
	ctx := context.Background()

	_ = c.Set(ctx, "k", "old", time.Minute)
	_ = c.Set(ctx, "k", "new", time.Minute)

	val, err := c.Get(ctx, "k")
	assert.NoError(t, err)
	assert.Equal(t, "new", val)
}

func TestInMemoryCache_Concurrent(t *testing.T) {
	c := NewInMemory()
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			_ = c.Set(ctx, "shared", "goroutine", time.Minute)
			_, _ = c.Get(ctx, "shared")
		}
		done <- struct{}{}
	}()

	for i := 0; i < 100; i++ {
		_ = c.Set(ctx, "shared", "main", time.Minute)
		_, _ = c.Get(ctx, "shared")
	}
	<-done

	val, err := c.Get(ctx, "shared")
	assert.NoError(t, err)
	assert.Contains(t, []string{"main", "goroutine"}, val)
}
