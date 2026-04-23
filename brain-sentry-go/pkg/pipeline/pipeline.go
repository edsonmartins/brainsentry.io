// Package pipeline implements a small, dependency-aware DAG runner inspired
// by semantica's PipelineBuilder. Steps declare their IDs and the IDs they
// depend on; the executor runs them concurrently when no dependency blocks,
// stops on first error, and returns the output map keyed by step ID.
package pipeline

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// StepFunc is the unit of work for a step. Inputs map contains results from
// upstream steps keyed by step ID.
type StepFunc func(ctx context.Context, inputs map[string]any) (any, error)

// Step is one DAG node.
type Step struct {
	ID        string
	DependsOn []string
	Run       StepFunc
}

// Pipeline is an immutable snapshot of steps built via Builder.Build.
type Pipeline struct {
	steps map[string]Step
	order []string // topological order
}

// Builder accumulates steps before validation and Build.
type Builder struct {
	steps []Step
}

// NewBuilder returns an empty Builder.
func NewBuilder() *Builder { return &Builder{} }

// Add registers a step; chains for readability.
func (b *Builder) Add(id string, run StepFunc, deps ...string) *Builder {
	b.steps = append(b.steps, Step{ID: id, DependsOn: deps, Run: run})
	return b
}

// Build returns a validated Pipeline or an error if the DAG is invalid.
func (b *Builder) Build() (*Pipeline, error) {
	stepsByID := make(map[string]Step, len(b.steps))
	for _, s := range b.steps {
		if s.ID == "" {
			return nil, fmt.Errorf("step has empty id")
		}
		if _, dup := stepsByID[s.ID]; dup {
			return nil, fmt.Errorf("duplicate step id: %s", s.ID)
		}
		if s.Run == nil {
			return nil, fmt.Errorf("step %q has nil Run", s.ID)
		}
		stepsByID[s.ID] = s
	}
	for _, s := range b.steps {
		for _, dep := range s.DependsOn {
			if _, ok := stepsByID[dep]; !ok {
				return nil, fmt.Errorf("step %q depends on unknown step %q", s.ID, dep)
			}
		}
	}
	order, err := topoSort(stepsByID)
	if err != nil {
		return nil, err
	}
	return &Pipeline{steps: stepsByID, order: order}, nil
}

// Result carries the per-step outputs and timing.
type Result struct {
	Outputs   map[string]any
	Durations map[string]time.Duration
	Order     []string
}

// Run executes the pipeline concurrently honouring dependencies. Stops on the
// first failing step (cancels the context) and returns partial results + error.
func (p *Pipeline) Run(ctx context.Context) (*Result, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	outputs := &sync.Map{}
	durations := &sync.Map{}
	errCh := make(chan error, 1)
	var firstErr sync.Once

	inDegree := make(map[string]int, len(p.steps))
	dependents := make(map[string][]string, len(p.steps))
	for id, s := range p.steps {
		inDegree[id] = len(s.DependsOn)
		for _, dep := range s.DependsOn {
			dependents[dep] = append(dependents[dep], id)
		}
	}

	var mu sync.Mutex
	var wg sync.WaitGroup

	var launch func(id string)
	launch = func(id string) {
		defer wg.Done()

		if ctx.Err() != nil {
			return
		}

		step := p.steps[id]
		inputs := make(map[string]any, len(step.DependsOn))
		for _, dep := range step.DependsOn {
			if v, ok := outputs.Load(dep); ok {
				inputs[dep] = v
			}
		}

		start := time.Now()
		out, err := step.Run(ctx, inputs)
		durations.Store(id, time.Since(start))
		if err != nil {
			firstErr.Do(func() {
				errCh <- fmt.Errorf("step %q failed: %w", id, err)
				cancel()
			})
			return
		}
		outputs.Store(id, out)

		mu.Lock()
		var nowReady []string
		for _, next := range dependents[id] {
			inDegree[next]--
			if inDegree[next] == 0 {
				nowReady = append(nowReady, next)
			}
		}
		mu.Unlock()

		for _, next := range nowReady {
			wg.Add(1)
			go launch(next)
		}
	}

	mu.Lock()
	var initial []string
	for id, n := range inDegree {
		if n == 0 {
			initial = append(initial, id)
		}
	}
	mu.Unlock()

	for _, id := range initial {
		wg.Add(1)
		go launch(id)
	}
	wg.Wait()

	result := &Result{
		Outputs:   make(map[string]any, len(p.steps)),
		Durations: make(map[string]time.Duration, len(p.steps)),
		Order:     p.order,
	}
	outputs.Range(func(k, v any) bool {
		result.Outputs[k.(string)] = v
		return true
	})
	durations.Range(func(k, v any) bool {
		result.Durations[k.(string)] = v.(time.Duration)
		return true
	})

	select {
	case err := <-errCh:
		return result, err
	default:
		return result, nil
	}
}

// topoSort returns a Kahn-style topological ordering of step IDs or an error
// when the DAG contains a cycle.
func topoSort(steps map[string]Step) ([]string, error) {
	inDegree := make(map[string]int, len(steps))
	for id, s := range steps {
		inDegree[id] += 0
		for _, dep := range s.DependsOn {
			inDegree[id]++
			_ = steps[dep]
		}
	}
	var queue []string
	for id, n := range inDegree {
		if n == 0 {
			queue = append(queue, id)
		}
	}
	order := make([]string, 0, len(steps))
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		order = append(order, id)
		for _, s := range steps {
			for _, dep := range s.DependsOn {
				if dep == id {
					inDegree[s.ID]--
					if inDegree[s.ID] == 0 {
						queue = append(queue, s.ID)
					}
				}
			}
		}
	}
	if len(order) != len(steps) {
		return nil, fmt.Errorf("cycle detected in pipeline")
	}
	return order, nil
}
