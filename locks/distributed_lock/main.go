package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	lredis "storage-journey/locks/distributed_lock/redis"
	"time"

	"github.com/redis/go-redis/v9"
)

func processPayment() error {
	log.Println("Running exclusive payment logic...")
	time.Sleep(2 * time.Second)
	return nil
}

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()

	ctx := context.Background()

	ownerID := fmt.Sprintf("worker-%d", time.Now().UnixNano())
	lock := lredis.NewRedisLock(rdb, "payment:lock:user-42", ownerID, 30*time.Second)

	log.Println("request 1")
	err := lock.Acquire(ctx)
	if errors.Is(err, lredis.ErrLockNotAcquired) {
		log.Println("Another process holds the lock, skipping.")
		return
	}
	if err != nil {
		log.Fatalf("unexpected error: %v", err)
	}
	defer lock.Release(ctx)
	log.Println("Lock acquired! Processing payment...")

	// Concurrent request
	log.Println("request 2")
	err = lock.Acquire(ctx)
	if errors.Is(err, lredis.ErrLockNotAcquired) {
		log.Println("Another process holds the lock, skipping.")
	}

	processPayment()
}
