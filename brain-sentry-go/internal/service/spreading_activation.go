package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"

	graphrepo "github.com/integraltech/brainsentry/internal/repository/graph"
)

// SpreadingActivationService propagates saliency boosts through graph neighbors.
// Simulates associative activation in cognitive science.
type SpreadingActivationService struct {
	graphRepo        *graphrepo.MemoryGraphRepository
	graphClient      *graphrepo.Client
	neighborProvider spreadingNeighborProvider
	maxHops          int     // max propagation depth
	decayFactor      float64 // decay per hop (0-1)
	minThreshold     float64 // stop propagation below this
}

type spreadingNeighborProvider interface {
	GetNeighbors(ctx context.Context, memoryID, tenantID string) ([]graphNeighbor, error)
}

// NewSpreadingActivationService creates a new SpreadingActivationService.
func NewSpreadingActivationService(
	graphRepo *graphrepo.MemoryGraphRepository,
	graphClient *graphrepo.Client,
) *SpreadingActivationService {
	var neighborProvider spreadingNeighborProvider
	if graphClient != nil {
		neighborProvider = graphClientNeighborProvider{client: graphClient}
	}
	return &SpreadingActivationService{
		graphRepo:        graphRepo,
		graphClient:      graphClient,
		neighborProvider: neighborProvider,
		maxHops:          3,
		decayFactor:      0.5,  // halves each hop
		minThreshold:     0.05, // stop below 5%
	}
}

// ActivationResult represents the spreading activation output.
type ActivationResult struct {
	SeedIDs        []string           `json:"seedIds"`
	Activations    []MemoryActivation `json:"activations"`
	TotalActivated int                `json:"totalActivated"`
	MaxHops        int                `json:"maxHops"`
}

// MemoryActivation represents a memory's activation level.
type MemoryActivation struct {
	MemoryID     string  `json:"memoryId"`
	Activation   float64 `json:"activation"` // 0-1 activation strength
	HopsFromSeed int     `json:"hopsFromSeed"`
	PathStrength float64 `json:"pathStrength"` // edge strength along path
}

// Spread propagates activation from seed memories through graph neighbors.
func (s *SpreadingActivationService) Spread(ctx context.Context, seedIDs []string, seedActivations []float64, tenantID string) (*ActivationResult, error) {
	if s.neighborProvider == nil || len(seedIDs) == 0 {
		return &ActivationResult{SeedIDs: seedIDs}, nil
	}

	result := &ActivationResult{
		SeedIDs: seedIDs,
		MaxHops: s.maxHops,
	}

	// Initialize activation map with seeds
	activationMap := make(map[string]float64)
	hopMap := make(map[string]int)
	strengthMap := make(map[string]float64)

	for i, id := range seedIDs {
		activation := 1.0
		if i < len(seedActivations) {
			activation = seedActivations[i]
		}
		activationMap[id] = activation
		hopMap[id] = 0
		strengthMap[id] = 1.0
	}

	// BFS propagation
	frontier := make([]string, len(seedIDs))
	copy(frontier, seedIDs)

	for hop := 1; hop <= s.maxHops; hop++ {
		var nextFrontier []string

		for _, nodeID := range frontier {
			currentActivation := activationMap[nodeID]
			propagated := currentActivation * s.decayFactor

			if propagated < s.minThreshold {
				continue
			}

			// Get neighbors from graph
			neighbors, err := s.getNeighbors(ctx, nodeID, tenantID)
			if err != nil {
				slog.Debug("failed to get neighbors", "id", nodeID, "error", err)
				continue
			}

			for _, neighbor := range neighbors {
				// Weight propagation by edge strength
				edgeWeight := math.Min(neighbor.Strength/10.0, 1.0)
				neighborActivation := propagated * edgeWeight

				if neighborActivation < s.minThreshold {
					continue
				}

				// Update if this path provides stronger activation
				existing, seen := activationMap[neighbor.ID]
				if !seen || neighborActivation > existing {
					activationMap[neighbor.ID] = neighborActivation
					hopMap[neighbor.ID] = hop
					strengthMap[neighbor.ID] = neighbor.Strength
					if !seen {
						nextFrontier = append(nextFrontier, neighbor.ID)
					}
				}
			}
		}

		frontier = nextFrontier
		if len(frontier) == 0 {
			break
		}
	}

	// Build results (exclude seeds)
	seedSet := make(map[string]bool, len(seedIDs))
	for _, id := range seedIDs {
		seedSet[id] = true
	}

	for id, activation := range activationMap {
		if seedSet[id] {
			continue
		}
		result.Activations = append(result.Activations, MemoryActivation{
			MemoryID:     id,
			Activation:   activation,
			HopsFromSeed: hopMap[id],
			PathStrength: strengthMap[id],
		})
	}

	// Sort by activation descending
	sortActivations(result.Activations)
	result.TotalActivated = len(result.Activations)

	return result, nil
}

type graphNeighbor struct {
	ID       string
	Strength float64
}

func (s *SpreadingActivationService) getNeighbors(ctx context.Context, memoryID, tenantID string) ([]graphNeighbor, error) {
	if s.neighborProvider == nil {
		return nil, nil
	}
	return s.neighborProvider.GetNeighbors(ctx, memoryID, tenantID)
}

type graphClientNeighborProvider struct {
	client *graphrepo.Client
}

func (p graphClientNeighborProvider) GetNeighbors(ctx context.Context, memoryID, tenantID string) ([]graphNeighbor, error) {
	cypher := fmt.Sprintf(`MATCH (m:Memory {id: '%s'})-[r:RELATED_TO]-(neighbor:Memory)
WHERE neighbor.tenantId = '%s'
RETURN neighbor.id as id, coalesce(r.strength, 1) as strength
LIMIT 20`,
		graphrepo.EscapeCypher(memoryID),
		graphrepo.EscapeCypher(tenantID),
	)

	result, err := p.client.Query(ctx, cypher)
	if err != nil {
		return nil, err
	}

	neighbors := make([]graphNeighbor, 0, len(result.Records))
	for _, rec := range result.Records {
		neighbors = append(neighbors, graphNeighbor{
			ID:       graphrepo.GetString(rec.Values, "id"),
			Strength: graphrepo.GetFloat64(rec.Values, "strength"),
		})
	}

	return neighbors, nil
}

func sortActivations(activations []MemoryActivation) {
	for i := 1; i < len(activations); i++ {
		key := activations[i]
		j := i - 1
		for j >= 0 && activations[j].Activation < key.Activation {
			activations[j+1] = activations[j]
			j--
		}
		activations[j+1] = key
	}
}
