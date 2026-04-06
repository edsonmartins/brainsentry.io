package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// LLMProvider abstracts LLM chat completion calls.
type LLMProvider interface {
	Name() string
	Chat(ctx context.Context, messages []ChatMessage) (string, error)
}

// OpenRouterProvider wraps OpenRouterService as an LLMProvider.
type OpenRouterProvider struct {
	service *OpenRouterService
}

func NewOpenRouterProvider(s *OpenRouterService) *OpenRouterProvider {
	return &OpenRouterProvider{service: s}
}

func (p *OpenRouterProvider) Name() string { return "openrouter" }

func (p *OpenRouterProvider) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	return p.service.Chat(ctx, messages)
}

// FallbackChainProvider tries providers sequentially until one succeeds.
// Integrates circuit breakers per provider to avoid wasting time on known-failing backends.
type FallbackChainProvider struct {
	providers []LLMProvider
	breakers  map[string]*CircuitBreaker
	mu        sync.RWMutex
}

// NewFallbackChainProvider creates a chain of LLM providers with circuit breakers.
func NewFallbackChainProvider(providers ...LLMProvider) *FallbackChainProvider {
	breakers := make(map[string]*CircuitBreaker, len(providers))
	for _, p := range providers {
		breakers[p.Name()] = NewCircuitBreaker(CircuitBreakerConfig{
			Name:             "llm-" + p.Name(),
			FailureThreshold: 3,
			SuccessThreshold: 1,
			OpenTimeout:      30 * time.Second,
			MaxRetries:       0, // single attempt per provider; chain handles retries
			BaseBackoff:      500 * time.Millisecond,
			MaxBackoff:       5 * time.Second,
		})
	}
	return &FallbackChainProvider{
		providers: providers,
		breakers:  breakers,
	}
}

func (f *FallbackChainProvider) Name() string { return "fallback-chain" }

// Chat tries each provider in order, skipping those with open circuit breakers.
func (f *FallbackChainProvider) Chat(ctx context.Context, messages []ChatMessage) (string, error) {
	var lastErr error

	for _, p := range f.providers {
		cb := f.breakers[p.Name()]

		var response string
		err := cb.Execute(ctx, func(ctx context.Context) error {
			var chatErr error
			response, chatErr = p.Chat(ctx, messages)
			return chatErr
		})

		if err != nil {
			slog.Warn("LLM provider failed, trying next",
				"provider", p.Name(),
				"error", err,
				"circuit_state", cb.State().String(),
			)
			lastErr = err
			continue
		}

		return response, nil
	}

	return "", fmt.Errorf("all LLM providers failed: %w", lastErr)
}

// ProviderStatus returns the circuit breaker state for each provider.
func (f *FallbackChainProvider) ProviderStatus() map[string]string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	status := make(map[string]string, len(f.providers))
	for _, p := range f.providers {
		status[p.Name()] = f.breakers[p.Name()].State().String()
	}
	return status
}
