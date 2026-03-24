package lredis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrLockNotAcquired = errors.New("lock not acquired")
	ErrLockNotOwned    = errors.New("lock not owned by this process")
)

type RedisLock struct {
	client *redis.Client
	key    string
	value  string
	ttl    time.Duration
}

func NewRedisLock(client *redis.Client, key, ownerID string, ttl time.Duration) *RedisLock {
	return &RedisLock{
		client: client,
		key:    key,
		value:  ownerID,
		ttl:    ttl,
	}
}

// Acquire tries to get the lock once. Returns ErrLockNotAcquired if busy.
func (l *RedisLock) Acquire(ctx context.Context) error {
	res, err := l.client.SetArgs(ctx, l.key, l.value, redis.SetArgs{
		Mode: string(redis.NX),
		TTL:  time.Second * 30,
	}).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}
	if res != "OK" {
		return ErrLockNotAcquired
	}

	return nil
}

// Release deletes the lock only if we own it (atomic Lua script).
func (l *RedisLock) Release(ctx context.Context) error {
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`
	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value).Int()
	if err != nil {
		return fmt.Errorf("redis error: %w", err)
	}
	if result == 0 {
		return ErrLockNotOwned
	}

	return nil
}

// AcquireWithRetry retries acquiring the lock until it succeeds or times out.
func (l *RedisLock) AcquireWithRetry(ctx context.Context, retryDelay time.Duration) error {
	for {
		err := l.Acquire(ctx)
		if err == nil {
			return nil // acquired!
		}
		if !errors.Is(err, ErrLockNotAcquired) {
			return err // real error, stop retrying
		}

		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for lock: %w", ctx.Err())
		case <-time.After(retryDelay):
			// try again
		}
	}
}

// WithLock acquires the lock, runs fn, then always releases the lock.
func WithLock(ctx context.Context, lock *RedisLock, fn func() error) error {
	if err := lock.Acquire(ctx); err != nil {
		return fmt.Errorf("acquire lock: %w", err)
	}
	defer func() {
		if err := lock.Release(ctx); err != nil {
			log.Printf("warning: failed to release lock: %v", err)
		}
	}()

	return fn()
}
