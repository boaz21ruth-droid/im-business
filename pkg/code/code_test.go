package code

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:16379", Password: "openIM123", DB: 15})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	t.Cleanup(func() { rdb.FlushDB(context.Background()) })
	return rdb
}

func TestSendAndVerifyOK(t *testing.T) {
	rdb := newTestRedis(t)
	svc := NewService(rdb)
	if err := svc.Send(context.Background(), "test@example.com"); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if err := svc.Verify(context.Background(), "test@example.com", "123456"); err != nil {
		t.Fatalf("Verify: %v", err)
	}
}

func TestVerifyWrongCode(t *testing.T) {
	rdb := newTestRedis(t)
	svc := NewService(rdb)
	svc.Send(context.Background(), "test@example.com")
	if err := svc.Verify(context.Background(), "test@example.com", "999999"); err == nil {
		t.Fatal("expected error for wrong code")
	}
}

func TestVerifyExpired(t *testing.T) {
	rdb := newTestRedis(t)
	svc := NewService(rdb)
	// no Send — key doesn't exist
	if err := svc.Verify(context.Background(), "ghost@example.com", "123456"); err == nil {
		t.Fatal("expected error for non-existent code")
	}
}
