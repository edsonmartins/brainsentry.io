package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"

	graphrepo "github.com/integraltech/brainsentry/internal/repository/graph"
)

// LouvainService detects communities in the memory graph using the Louvain method.
// Communities represent clusters of densely connected memories.
type LouvainService struct {
	graphClient   *graphrepo.Client
	maxIterations int
	minModularity float64 // stop if modularity gain < this
}

// NewLouvainService creates a new LouvainService.
func NewLouvainService(graphClient *graphrepo.Client) *LouvainService {
	return &LouvainService{
		graphClient:   graphClient,
		maxIterations: 10,
		minModularity: 0.001,
	}
}

// Community represents a detected community of memories.
type Community struct {
	ID        int      `json:"id"`
	MemberIDs []string `json:"memberIds"`
	Size      int      `json:"size"`
	Density   float64  `json:"density"` // internal edge density
}

// CommunityResult represents the output of community detection.
type CommunityResult struct {
	Communities []Community `json:"communities"`
	TotalNodes  int         `json:"totalNodes"`
	TotalEdges  int         `json:"totalEdges"`
	Modularity  float64     `json:"modularity"`
	Iterations  int         `json:"iterations"`
}

// graphEdge represents an edge in the graph for Louvain computation.
type graphEdge struct {
	From   string
	To     string
	Weight float64
}

// CommunityLink is the backend-independent weighted edge consumed by the
// community detector. It lets API views analyse the exact filtered graph
// being displayed, including canonical PostgreSQL relationships.
type CommunityLink struct {
	From   string
	To     string
	Weight float64
}

// DetectCommunities runs Louvain community detection on the memory graph for a tenant.
func (s *LouvainService) DetectCommunities(ctx context.Context, tenantID string) (*CommunityResult, error) {
	if s.graphClient == nil {
		return &CommunityResult{}, nil
	}

	// Fetch all edges for tenant
	edges, err := s.fetchEdges(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("fetching edges: %w", err)
	}

	links := make([]CommunityLink, 0, len(edges))
	for _, edge := range edges {
		links = append(links, CommunityLink(edge))
	}
	return s.DetectCommunitiesFromLinks(links), nil
}

// DetectCommunitiesFromLinks analyses exactly the supplied graph.
func (s *LouvainService) DetectCommunitiesFromLinks(edges []CommunityLink) *CommunityResult {
	if len(edges) == 0 {
		return &CommunityResult{}
	}

	// Build adjacency structure
	nodeSet := make(map[string]bool)
	for _, e := range edges {
		nodeSet[e.From] = true
		nodeSet[e.To] = true
	}

	nodes := make([]string, 0, len(nodeSet))
	nodeIndex := make(map[string]int)
	for n := range nodeSet {
		nodeIndex[n] = len(nodes)
		nodes = append(nodes, n)
	}

	n := len(nodes)

	// Build weighted adjacency list
	adj := make([][]weightedNeighbor, n)
	for i := range adj {
		adj[i] = make([]weightedNeighbor, 0)
	}

	totalWeight := 0.0
	for _, e := range edges {
		i, j := nodeIndex[e.From], nodeIndex[e.To]
		w := e.Weight
		if w <= 0 {
			w = 1.0
		}
		adj[i] = append(adj[i], weightedNeighbor{node: j, weight: w})
		adj[j] = append(adj[j], weightedNeighbor{node: i, weight: w})
		totalWeight += w
	}

	if totalWeight == 0 {
		return &CommunityResult{TotalNodes: n, TotalEdges: len(edges)}
	}

	// Compute node strengths (sum of edge weights)
	strength := make([]float64, n)
	for i := 0; i < n; i++ {
		for _, nb := range adj[i] {
			strength[i] += nb.weight
		}
	}

	// Initialize: each node in its own community
	community := make([]int, n)
	for i := range community {
		community[i] = i
	}

	m2 := 2.0 * totalWeight
	bestModularity := computeModularity(community, adj, strength, m2, n)

	// Phase 1: local moves. Evaluate the real modularity for every candidate
	// community. This is deliberately exact: the previous shortcut omitted
	// non-adjacent and diagonal pairs and could report impossible values such
	// as -9 for a graph with no visible communities.
	var iterations int
	for iter := 0; iter < s.maxIterations; iter++ {
		iterations++
		improved := false

		for i := 0; i < n; i++ {
			currentComm := community[i]
			candidateCommunities := make(map[int]bool)
			for _, nb := range adj[i] {
				candidateCommunities[community[nb.node]] = true
			}
			bestComm := currentComm
			currentModularity := computeModularity(community, adj, strength, m2, n)
			bestCandidateModularity := currentModularity
			for c := range candidateCommunities {
				if c == currentComm {
					continue
				}
				community[i] = c
				candidateModularity := computeModularity(community, adj, strength, m2, n)
				if candidateModularity > bestCandidateModularity {
					bestCandidateModularity = candidateModularity
					bestComm = c
				}
			}
			community[i] = currentComm
			if bestComm != currentComm && bestCandidateModularity-currentModularity > s.minModularity {
				community[i] = bestComm
				improved = true
			}
		}

		if !improved {
			break
		}

		bestModularity = computeModularity(community, adj, strength, m2, n)
	}

	// Build result communities
	commMembers := make(map[int][]string)
	for i, c := range community {
		commMembers[c] = append(commMembers[c], nodes[i])
	}

	communities := make([]Community, 0, len(commMembers))
	commID := 0
	for _, members := range commMembers {
		if len(members) < 2 {
			continue // skip singletons
		}
		density := computeCommunityDensity(members, nodeIndex, adj)
		communities = append(communities, Community{
			ID:        commID,
			MemberIDs: members,
			Size:      len(members),
			Density:   density,
		})
		commID++
	}

	// Sort by size descending
	sortCommunities(communities)

	result := &CommunityResult{
		Communities: communities,
		TotalNodes:  n,
		TotalEdges:  len(edges),
		Modularity:  bestModularity,
		Iterations:  iterations,
	}

	slog.Debug("Louvain community detection completed",
		"communities", len(communities),
		"nodes", n,
		"edges", len(edges),
		"modularity", bestModularity,
		"iterations", iterations,
	)

	return result
}

type weightedNeighbor struct {
	node   int
	weight float64
}

func (s *LouvainService) fetchEdges(ctx context.Context, tenantID string) ([]graphEdge, error) {
	cypher := fmt.Sprintf(`MATCH (a:Memory)-[r:RELATED_TO]->(b:Memory)
WHERE a.tenantId = '%s' AND b.tenantId = '%s'
RETURN a.id as fromId, b.id as toId, coalesce(r.strength, 1) as weight
LIMIT 1000`,
		graphrepo.EscapeCypher(tenantID),
		graphrepo.EscapeCypher(tenantID),
	)

	result, err := s.graphClient.Query(ctx, cypher)
	if err != nil {
		return nil, err
	}

	edges := make([]graphEdge, 0, len(result.Records))
	for _, rec := range result.Records {
		edges = append(edges, graphEdge{
			From:   graphrepo.GetString(rec.Values, "fromId"),
			To:     graphrepo.GetString(rec.Values, "toId"),
			Weight: graphrepo.GetFloat64(rec.Values, "weight"),
		})
	}

	return edges, nil
}

// computeModularity calculates Q = sum_c[L_c/m - (K_c/2m)^2].
func computeModularity(community []int, adj [][]weightedNeighbor, strength []float64, m2 float64, n int) float64 {
	if m2 == 0 {
		return 0
	}

	totalStrength := make(map[int]float64)
	internalDirectedWeight := make(map[int]float64)
	for i := 0; i < n; i++ {
		totalStrength[community[i]] += strength[i]
		for _, nb := range adj[i] {
			if community[i] == community[nb.node] {
				internalDirectedWeight[community[i]] += nb.weight
			}
		}
	}
	q := 0.0
	for c, k := range totalStrength {
		q += internalDirectedWeight[c]/m2 - math.Pow(k/m2, 2)
	}
	return q
}

func computeCommunityDensity(members []string, nodeIndex map[string]int, adj [][]weightedNeighbor) float64 {
	if len(members) < 2 {
		return 0
	}

	memberSet := make(map[int]bool, len(members))
	for _, m := range members {
		memberSet[nodeIndex[m]] = true
	}

	internalEdges := 0
	for _, m := range members {
		idx := nodeIndex[m]
		for _, nb := range adj[idx] {
			if memberSet[nb.node] {
				internalEdges++
			}
		}
	}
	internalEdges /= 2 // undirected

	maxEdges := len(members) * (len(members) - 1) / 2
	if maxEdges == 0 {
		return 0
	}

	return math.Min(float64(internalEdges)/float64(maxEdges), 1.0)
}

func sortCommunities(communities []Community) {
	for i := 1; i < len(communities); i++ {
		key := communities[i]
		j := i - 1
		for j >= 0 && communities[j].Size < key.Size {
			communities[j+1] = communities[j]
			j--
		}
		communities[j+1] = key
	}
}
