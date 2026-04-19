package service

import (
	"context"
	"testing"
	"time"
)

func TestSessionCache_PushAndList(t *testing.T) {
	backend := NewInMemorySessionBackend(10)
	cache := NewSessionMemoryCache(backend, DefaultSessionCacheConfig(), nil)

	ctx := context.Background()
	err := cache.Push(ctx, "s1", SessionInteraction{
		Query:    "What is Go?",
		Response: "A statically-typed language.",
	})
	if err != nil {
		t.Fatalf("push: %v", err)
	}

	items, err := cache.Recent(ctx, "s1", 10)
	if err != nil {
		t.Fatalf("recent: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
	if items[0].Query != "What is Go?" {
		t.Errorf("unexpected query: %s", items[0].Query)
	}
	if items[0].ID == "" {
		t.Error("expected auto-generated ID")
	}
	if items[0].CreatedAt.IsZero() {
		t.Error("expected auto-generated CreatedAt")
	}
}

func TestSessionCache_RequiresSessionID(t *testing.T) {
	cache := NewSessionMemoryCache(NewInMemorySessionBackend(10), DefaultSessionCacheConfig(), nil)
	err := cache.Push(context.Background(), "", SessionInteraction{Query: "x"})
	if err == nil {
		t.Error("expected error for empty sessionID")
	}
}

func TestSessionCache_NewestFirst(t *testing.T) {
	backend := NewInMemorySessionBackend(10)
	cache := NewSessionMemoryCache(backend, DefaultSessionCacheConfig(), nil)

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		cache.Push(ctx, "s1", SessionInteraction{Query: string(rune('A' + i))})
		time.Sleep(time.Millisecond) // guarantee distinct timestamps
	}

	items, _ := cache.Recent(ctx, "s1", 10)
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
	if items[0].Query != "C" {
		t.Errorf("newest should be C, got %s", items[0].Query)
	}
}

func TestSessionCache_Clear(t *testing.T) {
	backend := NewInMemorySessionBackend(10)
	cache := NewSessionMemoryCache(backend, DefaultSessionCacheConfig(), nil)

	ctx := context.Background()
	cache.Push(ctx, "s1", SessionInteraction{Query: "x"})
	cache.Clear(ctx, "s1")

	items, _ := cache.Recent(ctx, "s1", 10)
	if len(items) != 0 {
		t.Errorf("expected 0 items after clear, got %d", len(items))
	}
}

func TestSessionCache_MaxLen(t *testing.T) {
	backend := NewInMemorySessionBackend(3)
	cache := NewSessionMemoryCache(backend, DefaultSessionCacheConfig(), nil)

	ctx := context.Background()
	for i := 0; i < 5; i++ {
		cache.Push(ctx, "s1", SessionInteraction{Query: "q"})
	}

	items, _ := cache.Recent(ctx, "s1", 100)
	if len(items) != 3 {
		t.Errorf("expected cap at 3, got %d", len(items))
	}
}

func TestSessionCache_ListSessions(t *testing.T) {
	backend := NewInMemorySessionBackend(10)
	cache := NewSessionMemoryCache(backend, DefaultSessionCacheConfig(), nil)

	ctx := context.Background()
	cache.Push(ctx, "s1", SessionInteraction{Query: "a"})
	cache.Push(ctx, "s2", SessionInteraction{Query: "b"})

	keys, err := cache.ListSessions(ctx)
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(keys) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(keys))
	}
}

func TestSessionCache_CognifyWithoutService(t *testing.T) {
	cache := NewSessionMemoryCache(NewInMemorySessionBackend(10), DefaultSessionCacheConfig(), nil)

	_, err := cache.Cognify(context.Background(), "s1", false)
	if err == nil {
		t.Error("expected error when memory service is nil")
	}
}

func TestSessionCache_PreservesProvidedFields(t *testing.T) {
	backend := NewInMemorySessionBackend(10)
	cache := NewSessionMemoryCache(backend, DefaultSessionCacheConfig(), nil)

	providedID := "fixed-id"
	providedTime := time.Now().Add(-time.Hour)

	cache.Push(context.Background(), "s1", SessionInteraction{
		ID:        providedID,
		CreatedAt: providedTime,
		Query:     "x",
	})

	items, _ := cache.Recent(context.Background(), "s1", 10)
	if items[0].ID != providedID {
		t.Errorf("expected ID preserved, got %s", items[0].ID)
	}
	if !items[0].CreatedAt.Equal(providedTime) {
		t.Errorf("expected CreatedAt preserved")
	}
}
