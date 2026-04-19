package graph

import "context"

// GraphBackend abstracts the minimal contract needed by MemoryGraphRepository,
// EntityGraphRepository, and GraphRAGRepository to talk to a graph database.
//
// The FalkorDB Client is the current default implementation. This interface
// is designed to allow swapping in Neo4j, Kuzu, or other backends in the
// future without touching the service/repository layer.
//
// Implementations must be safe for concurrent use.
type GraphBackend interface {
	// Query executes a Cypher query and returns the result set.
	// Different backends may have slightly different Cypher dialects;
	// implementations are responsible for dialect adaptation if needed.
	Query(ctx context.Context, cypher string) (*QueryResult, error)

	// Close releases any underlying resources (connection pools, etc.).
	Close() error

	// Name returns the implementation identifier (e.g., "falkordb", "neo4j").
	Name() string
}

// Compile-time check: the existing Client satisfies GraphBackend.
var _ GraphBackend = (*Client)(nil)

// Name returns "falkordb" for the FalkorDB-backed Client.
func (c *Client) Name() string { return "falkordb" }
