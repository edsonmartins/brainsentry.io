// Package lazy provides a generic, thread-safe lazy initializer analogous to
// semantica's _ModuleProxy. Expensive resources (graph clients, LLM providers,
// remote caches) are constructed only when first requested, shrinking cold
// start time and side-stepping optional backends in minimal environments.
package lazy

import (
	"sync"
	"sync/atomic"
)

// Value holds a lazily constructed T. The constructor is invoked at most once
// and its result (value + error) is cached for all subsequent Get calls.
type Value[T any] struct {
	once sync.Once
	done atomic.Bool
	ctor func() (T, error)
	val  T
	err  error
}

// New wraps the constructor; nothing runs until Get is called.
func New[T any](ctor func() (T, error)) *Value[T] {
	return &Value[T]{ctor: ctor}
}

// Get returns the cached value and error, constructing on the first call.
func (v *Value[T]) Get() (T, error) {
	v.once.Do(func() {
		v.val, v.err = v.ctor()
		v.done.Store(true)
	})
	return v.val, v.err
}

// MustGet panics if construction failed. Useful in boot paths where the
// caller has already decided the resource is mandatory.
func (v *Value[T]) MustGet() T {
	val, err := v.Get()
	if err != nil {
		panic(err)
	}
	return val
}

// Initialized reports whether the constructor has already run. Non-triggering.
func (v *Value[T]) Initialized() bool {
	return v.done.Load()
}
