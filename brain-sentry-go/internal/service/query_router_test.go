package service

import "testing"

func newRouter() *QueryRouterService {
	return NewQueryRouterService(DefaultQueryRouterConfig())
}

func TestQueryRouter_Coding(t *testing.T) {
	r := newRouter()
	d := r.Classify("Why does the api endpoint return a 500 error?")

	if d.Strategy != StrategyCoding {
		t.Errorf("expected CODING, got %s (scores=%v)", d.Strategy, d.Scores)
	}
	if d.Fallback {
		t.Error("should not fallback for clear coding query")
	}
}

func TestQueryRouter_Temporal(t *testing.T) {
	r := newRouter()
	d := r.Classify("What did we decide last week?")

	if d.Strategy != StrategyTemporal {
		t.Errorf("expected TEMPORAL, got %s (scores=%v)", d.Strategy, d.Scores)
	}
}

func TestQueryRouter_Cypher(t *testing.T) {
	r := newRouter()
	d := r.Classify("MATCH (n:Memory) RETURN n LIMIT 5")

	if d.Strategy != StrategyCypher {
		t.Errorf("expected CYPHER, got %s", d.Strategy)
	}
}

func TestQueryRouter_Graph(t *testing.T) {
	r := newRouter()
	d := r.Classify("How does PostgreSQL depend on other services?")

	if d.Strategy != StrategyGraph {
		t.Errorf("expected GRAPH, got %s (scores=%v)", d.Strategy, d.Scores)
	}
}

func TestQueryRouter_Entity(t *testing.T) {
	r := newRouter()
	d := r.Classify(`Who is "John Doe"?`)

	if d.Strategy != StrategyEntity {
		t.Errorf("expected ENTITY, got %s (scores=%v)", d.Strategy, d.Scores)
	}
}

func TestQueryRouter_NegationSkipped(t *testing.T) {
	r := newRouter()
	// "not related to" should NOT trigger GRAPH strategy
	d := r.Classify("Things that are not related to databases")

	// Should fallback (no strong match remaining)
	scores := d.Scores
	if scores[StrategyGraph] > 0 {
		t.Errorf("negation should have suppressed GRAPH score, got %v", scores)
	}
}

func TestQueryRouter_Empty(t *testing.T) {
	r := newRouter()
	d := r.Classify("")

	if !d.Fallback {
		t.Error("empty query should fallback")
	}
	if d.Strategy != StrategyHybrid {
		t.Errorf("expected HYBRID fallback, got %s", d.Strategy)
	}
}

func TestQueryRouter_Fallback(t *testing.T) {
	r := newRouter()
	// Vague query with no patterns
	d := r.Classify("okay sure")

	if !d.Fallback {
		t.Error("vague query should fallback")
	}
}

func TestQueryRouter_CodeFile(t *testing.T) {
	r := newRouter()
	d := r.Classify("Show me changes in main.go")

	if d.Strategy != StrategyCoding {
		t.Errorf("expected CODING for .go file mention, got %s", d.Strategy)
	}
}

func TestQueryRouter_ConfidenceInRange(t *testing.T) {
	r := newRouter()
	d := r.Classify("MATCH (n) RETURN n LIMIT 5")

	if d.Confidence < 0 || d.Confidence > 1 {
		t.Errorf("confidence should be in [0,1], got %f", d.Confidence)
	}
}

func TestQueryRouter_Semantic(t *testing.T) {
	r := newRouter()
	d := r.Classify("Anything about concepts similar to neural networks")

	if d.Strategy != StrategySemantic {
		t.Errorf("expected SEMANTIC, got %s (scores=%v)", d.Strategy, d.Scores)
	}
}
