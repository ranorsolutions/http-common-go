package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestCache(t *testing.T) (Cache, func()) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	c := NewRedisCache(client, 1*time.Minute)
	return c, mr.Close
}

func TestSetAndGetJSON(t *testing.T) {
	cache, cleanup := newTestCache(t)
	defer cleanup()

	type user struct {
		Name string
		Age  int
	}
	ctx := context.Background()
	u := user{"Alice", 30}

	err := cache.SetJSON(ctx, "user:1", u, 0)
	if err != nil {
		t.Fatalf("SetJSON failed: %v", err)
	}

	var out user
	found, err := cache.GetJSON(ctx, "user:1", &out)
	if err != nil || !found {
		t.Fatalf("GetJSON failed: found=%v err=%v", found, err)
	}
	if out != u {
		t.Errorf("expected %+v, got %+v", u, out)
	}
}

func TestDelete(t *testing.T) {
	cache, cleanup := newTestCache(t)
	defer cleanup()

	ctx := context.Background()
	cache.SetJSON(ctx, "k", "v", 0)

	if err := cache.Delete(ctx, "k"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	var out string
	found, _ := cache.GetJSON(ctx, "k", &out)
	if found {
		t.Error("expected key to be deleted")
	}
}
