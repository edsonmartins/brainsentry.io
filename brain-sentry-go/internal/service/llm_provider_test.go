package service

import (
	"context"
	"errors"
	"testing"
)

// mockLLMProvider implements LLMProvider for testing.
type mockLLMProvider struct {
	name      string
	response  string
	err       error
	callCount int
}

func (m *mockLLMProvider) Name() string { return m.name }

func (m *mockLLMProvider) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	m.callCount++
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func TestFallbackChainProvider_FirstSucceeds(t *testing.T) {
	p1 := &mockLLMProvider{name: "primary", response: "hello from primary"}
	p2 := &mockLLMProvider{name: "secondary", response: "hello from secondary"}

	chain := NewFallbackChainProvider(p1, p2)
	resp, err := chain.Chat(context.Background(), []ChatMessage{{Role: "user", Content: "test"}})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "hello from primary" {
		t.Errorf("expected primary response, got: %s", resp)
	}
	if p1.callCount != 1 {
		t.Errorf("expected primary called once, got %d", p1.callCount)
	}
	if p2.callCount != 0 {
		t.Errorf("expected secondary not called, got %d", p2.callCount)
	}
}

func TestFallbackChainProvider_FallsBackToSecond(t *testing.T) {
	p1 := &mockLLMProvider{name: "primary", err: errors.New("provider down")}
	p2 := &mockLLMProvider{name: "secondary", response: "hello from secondary"}

	chain := NewFallbackChainProvider(p1, p2)
	resp, err := chain.Chat(context.Background(), []ChatMessage{{Role: "user", Content: "test"}})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "hello from secondary" {
		t.Errorf("expected secondary response, got: %s", resp)
	}
	if p1.callCount != 1 {
		t.Errorf("expected primary called once, got %d", p1.callCount)
	}
	if p2.callCount != 1 {
		t.Errorf("expected secondary called once, got %d", p2.callCount)
	}
}

func TestFallbackChainProvider_AllFail(t *testing.T) {
	p1 := &mockLLMProvider{name: "primary", err: errors.New("down")}
	p2 := &mockLLMProvider{name: "secondary", err: errors.New("also down")}

	chain := NewFallbackChainProvider(p1, p2)
	_, err := chain.Chat(context.Background(), []ChatMessage{{Role: "user", Content: "test"}})

	if err == nil {
		t.Fatal("expected error when all providers fail")
	}
}

func TestFallbackChainProvider_Name(t *testing.T) {
	chain := NewFallbackChainProvider()
	if chain.Name() != "fallback-chain" {
		t.Errorf("expected name 'fallback-chain', got: %s", chain.Name())
	}
}

func TestFallbackChainProvider_ProviderStatus(t *testing.T) {
	p1 := &mockLLMProvider{name: "primary", response: "ok"}
	chain := NewFallbackChainProvider(p1)

	status := chain.ProviderStatus()
	if status["primary"] != "CLOSED" {
		t.Errorf("expected CLOSED, got: %s", status["primary"])
	}
}

func TestOpenRouterProvider_Name(t *testing.T) {
	p := NewOpenRouterProvider(&OpenRouterService{})
	if p.Name() != "openrouter" {
		t.Errorf("expected 'openrouter', got: %s", p.Name())
	}
}
