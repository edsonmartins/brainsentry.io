package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/integraltech/brainsentry/internal/domain"
	graphrepo "github.com/integraltech/brainsentry/internal/repository/graph"
	"github.com/integraltech/brainsentry/internal/repository/postgres"
	"github.com/integraltech/brainsentry/internal/service"
	"github.com/integraltech/brainsentry/pkg/tenant"
)

// GraphViewHandler exposes graph visualisation endpoints:
//   - GET /v1/graph/global   — full-tenant map with Louvain communities
//   - GET /v1/graph/ego      — ego-graph around a seed memory (multi-hop)
//   - GET /v1/graph/timeline — bi-temporal view with SUPERSEDES edges
type GraphViewHandler struct {
	memRepo     *postgres.MemoryRepository
	relRepo     *postgres.RelationshipRepository
	graphClient *graphrepo.Client
	graphRAG    *graphrepo.GraphRAGRepository
	louvain     *service.LouvainService
}

// NewGraphViewHandler constructs the handler.
func NewGraphViewHandler(
	memRepo *postgres.MemoryRepository,
	relRepo *postgres.RelationshipRepository,
	graphClient *graphrepo.Client,
	graphRAG *graphrepo.GraphRAGRepository,
	louvain *service.LouvainService,
) *GraphViewHandler {
	return &GraphViewHandler{
		memRepo:     memRepo,
		relRepo:     relRepo,
		graphClient: graphClient,
		graphRAG:    graphRAG,
		louvain:     louvain,
	}
}

// GraphNode is the node DTO shared by all three endpoints.
type GraphNode struct {
	ID              string     `json:"id"`
	MemoryID        string     `json:"memoryId,omitempty"`
	Label           string     `json:"label"`
	Category        string     `json:"category,omitempty"`
	Importance      string     `json:"importance,omitempty"`
	CommunityID     int        `json:"communityId"`
	AccessCount     int        `json:"accessCount,omitempty"`
	HelpfulCount    int        `json:"helpfulCount,omitempty"`
	NotHelpfulCount int        `json:"notHelpfulCount,omitempty"`
	EmotionalWeight float64    `json:"emotionalWeight,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	ValidFrom       *time.Time `json:"validFrom,omitempty"`
	ValidTo         *time.Time `json:"validTo,omitempty"`
	RecordedAt      time.Time  `json:"recordedAt"`
	SupersededBy    string     `json:"supersededBy,omitempty"`
	Tags            []string   `json:"tags,omitempty"`
	HopDistance     int        `json:"hopDistance,omitempty"`
	Score           float64    `json:"score,omitempty"`
	Version         int        `json:"version,omitempty"`
	Operation       string     `json:"operation,omitempty"`
	SystemFrom      *time.Time `json:"systemFrom,omitempty"`
	SystemTo        *time.Time `json:"systemTo,omitempty"`
}

// GraphEdge is the edge DTO.
type GraphEdge struct {
	Source   string  `json:"source"`
	Target   string  `json:"target"`
	Type     string  `json:"type,omitempty"`
	Strength float64 `json:"strength,omitempty"`
}

// GraphResponse is the common response envelope.
type GraphResponse struct {
	Nodes            []GraphNode         `json:"nodes"`
	Edges            []GraphEdge         `json:"edges"`
	Communities      []service.Community `json:"communities,omitempty"`
	Modularity       float64             `json:"modularity,omitempty"`
	TenantID         string              `json:"tenantId,omitempty"`
	Total            int                 `json:"total"`
	ProjectionStatus string              `json:"projectionStatus,omitempty"`
	Warnings         []string            `json:"warnings,omitempty"`
}

// Global handles GET /v1/graph/global
func (h *GraphViewHandler) Global(w http.ResponseWriter, r *http.Request) {
	if h.memRepo == nil {
		writeError(w, http.StatusServiceUnavailable, "memory repository not available")
		return
	}
	tenantID := tenant.FromContext(r.Context())

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 300
	}
	if limit > 2000 {
		limit = 2000
	}

	category := strings.TrimSpace(r.URL.Query().Get("category"))
	importance := strings.TrimSpace(r.URL.Query().Get("importance"))
	withCommunities := r.URL.Query().Get("communities") != "false"

	var memories []domain.Memory
	var err error
	switch {
	case category != "" && importance != "":
		memories, err = h.memRepo.FindByCategory(r.Context(), domain.MemoryCategory(category))
		if err == nil {
			filtered := memories[:0]
			for _, memory := range memories {
				if string(memory.Importance) == importance {
					filtered = append(filtered, memory)
				}
			}
			memories = filtered
		}
	case category != "":
		memories, err = h.memRepo.FindByCategory(r.Context(), domain.MemoryCategory(category))
	case importance != "":
		memories, err = h.memRepo.FindByImportance(r.Context(), domain.ImportanceLevel(importance))
	default:
		memories, _, err = h.memRepo.List(r.Context(), 0, limit)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "loading memories: "+err.Error())
		return
	}
	if len(memories) > limit {
		memories = memories[:limit]
	}

	memberSet := make(map[string]bool, len(memories))
	for _, m := range memories {
		memberSet[m.ID] = true
	}

	edges := make([]GraphEdge, 0)
	edgeSet := make(map[string]bool)
	addEdge := func(edge GraphEdge) {
		key := edge.Source + "→" + edge.Target + ":" + edge.Type
		if edge.Source == "" || edge.Target == "" || edgeSet[key] || !memberSet[edge.Source] || !memberSet[edge.Target] {
			return
		}
		edgeSet[key] = true
		edges = append(edges, edge)
	}
	if h.relRepo != nil {
		if relationships, relErr := h.relRepo.ListByTenant(r.Context()); relErr == nil {
			for _, relationship := range relationships {
				addEdge(GraphEdge{Source: relationship.FromMemoryID, Target: relationship.ToMemoryID, Type: string(relationship.Type), Strength: relationship.Strength})
			}
		} else {
			writeError(w, http.StatusInternalServerError, "loading canonical relationships: "+relErr.Error())
			return
		}
	}
	projectionStatus := "not_configured"
	warnings := make([]string, 0)
	if h.graphClient != nil {
		projectionStatus = "empty"
		cypher := fmt.Sprintf(`MATCH (a:Memory)-[r:RELATED_TO]->(b:Memory)
WHERE a.tenantId = '%s' AND b.tenantId = '%s'
RETURN a.id AS src, b.id AS tgt, coalesce(r.strength, 1.0) AS strength
LIMIT %d`, graphrepo.EscapeCypher(tenantID), graphrepo.EscapeCypher(tenantID), limit*10)
		if result, qerr := h.graphClient.Query(r.Context(), cypher); qerr == nil {
			for _, rec := range result.Records {
				src := graphrepo.GetString(rec.Values, "src")
				tgt := graphrepo.GetString(rec.Values, "tgt")
				if !memberSet[src] || !memberSet[tgt] {
					continue
				}
				addEdge(GraphEdge{
					Source:   src,
					Target:   tgt,
					Type:     "RELATED_TO",
					Strength: graphrepo.GetFloat64(rec.Values, "strength"),
				})
			}
			if len(edges) > 0 {
				projectionStatus = "ready"
			}
		} else {
			projectionStatus = "unavailable"
			warnings = append(warnings, "memory graph projection is unavailable")
		}
	}

	for _, m := range memories {
		if m.SupersededBy != "" && memberSet[m.SupersededBy] {
			addEdge(GraphEdge{
				Source:   m.ID,
				Target:   m.SupersededBy,
				Type:     "SUPERSEDES",
				Strength: 1.0,
			})
		}
	}

	communityMap := make(map[string]int)
	resp := GraphResponse{Total: len(memories), TenantID: tenantID, ProjectionStatus: projectionStatus, Warnings: warnings}
	if withCommunities && h.louvain != nil {
		links := make([]service.CommunityLink, 0, len(edges))
		for _, edge := range edges {
			if edge.Type == "SUPERSEDES" {
				continue
			}
			links = append(links, service.CommunityLink{From: edge.Source, To: edge.Target, Weight: edge.Strength})
		}
		cr := h.louvain.DetectCommunitiesFromLinks(links)
		for _, c := range cr.Communities {
			for _, mid := range c.MemberIDs {
				communityMap[mid] = c.ID
			}
		}
		resp.Communities = cr.Communities
		resp.Modularity = cr.Modularity
	}

	nodes := make([]GraphNode, 0, len(memories))
	for _, m := range memories {
		cid := -1
		if v, ok := communityMap[m.ID]; ok {
			cid = v
		}
		nodes = append(nodes, GraphNode{
			ID:              m.ID,
			MemoryID:        m.ID,
			Label:           graphLabel(m.Summary, m.Content, 120),
			Category:        string(m.Category),
			Importance:      string(m.Importance),
			CommunityID:     cid,
			AccessCount:     m.AccessCount,
			HelpfulCount:    m.HelpfulCount,
			NotHelpfulCount: m.NotHelpfulCount,
			EmotionalWeight: m.EmotionalWeight,
			CreatedAt:       m.CreatedAt,
			ValidFrom:       m.ValidFrom,
			ValidTo:         m.ValidTo,
			RecordedAt:      m.RecordedAt,
			SupersededBy:    m.SupersededBy,
			Tags:            m.Tags,
		})
	}

	resp.Nodes = nodes
	resp.Edges = edges
	writeJSON(w, http.StatusOK, resp)
}

// Ego handles GET /v1/graph/ego?memoryId=X&hops=2&limit=30
func (h *GraphViewHandler) Ego(w http.ResponseWriter, r *http.Request) {
	memoryID := strings.TrimSpace(r.URL.Query().Get("memoryId"))
	if memoryID == "" {
		writeError(w, http.StatusBadRequest, "memoryId is required")
		return
	}
	hops, _ := strconv.Atoi(r.URL.Query().Get("hops"))
	if hops <= 0 {
		hops = 2
	}
	if hops > 4 {
		hops = 4
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}

	tenantID := tenant.FromContext(r.Context())

	seed, err := h.memRepo.FindByID(r.Context(), memoryID)
	if err != nil || seed == nil {
		writeError(w, http.StatusNotFound, "seed memory not found")
		return
	}
	var results []graphrepo.MultiHopResult
	projectionStatus := "not_configured"
	warnings := make([]string, 0)
	if h.graphRAG != nil {
		projectionStatus = "empty"
		results, err = h.graphRAG.MultiHopSearch(r.Context(), []string{memoryID}, hops, limit, tenantID)
		if err != nil {
			projectionStatus = "unavailable"
			warnings = append(warnings, "derived graph projection is unavailable; canonical relationships are still included")
			results = nil
		} else if len(results) > 0 {
			projectionStatus = "ready"
		}
	}

	seen := map[string]bool{seed.ID: true}
	nodes := []GraphNode{{
		ID:              seed.ID,
		MemoryID:        seed.ID,
		Label:           graphLabel(seed.Summary, seed.Content, 120),
		Category:        string(seed.Category),
		Importance:      string(seed.Importance),
		CommunityID:     -1,
		AccessCount:     seed.AccessCount,
		HelpfulCount:    seed.HelpfulCount,
		NotHelpfulCount: seed.NotHelpfulCount,
		EmotionalWeight: seed.EmotionalWeight,
		CreatedAt:       seed.CreatedAt,
		ValidFrom:       seed.ValidFrom,
		ValidTo:         seed.ValidTo,
		RecordedAt:      seed.RecordedAt,
		SupersededBy:    seed.SupersededBy,
		Tags:            seed.Tags,
		HopDistance:     0,
		Score:           1.0,
	}}

	edgeSet := make(map[string]bool)
	edges := make([]GraphEdge, 0)
	addEdge := func(src, tgt string) {
		if src == "" || tgt == "" || src == tgt {
			return
		}
		key := src + "→" + tgt
		if edgeSet[key] {
			return
		}
		edgeSet[key] = true
		edges = append(edges, GraphEdge{Source: src, Target: tgt, Type: "RELATED_TO", Strength: 1.0})
	}

	for _, res := range results {
		if !seen[res.MemoryID] {
			nodes = append(nodes, GraphNode{
				ID:          res.MemoryID,
				MemoryID:    res.MemoryID,
				Label:       res.Summary,
				Category:    res.Category,
				Importance:  res.Importance,
				CommunityID: -1,
				HopDistance: res.HopDistance,
				Score:       res.Score,
			})
			seen[res.MemoryID] = true
		}
		for i := 0; i+1 < len(res.Path); i++ {
			addEdge(res.Path[i], res.Path[i+1])
		}
	}

	if h.relRepo != nil {
		relationships, relErr := h.relRepo.ListByTenant(r.Context())
		if relErr != nil {
			writeError(w, http.StatusInternalServerError, relErr.Error())
			return
		}
		type neighbor struct {
			id       string
			relType  string
			strength float64
		}
		adjacency := make(map[string][]neighbor)
		for _, relationship := range relationships {
			adjacency[relationship.FromMemoryID] = append(adjacency[relationship.FromMemoryID], neighbor{relationship.ToMemoryID, string(relationship.Type), relationship.Strength})
			adjacency[relationship.ToMemoryID] = append(adjacency[relationship.ToMemoryID], neighbor{relationship.FromMemoryID, string(relationship.Type), relationship.Strength})
		}
		distance := map[string]int{memoryID: 0}
		queue := []string{memoryID}
		for len(queue) > 0 && len(nodes) < limit {
			current := queue[0]
			queue = queue[1:]
			if distance[current] >= hops {
				continue
			}
			for _, next := range adjacency[current] {
				if _, exists := distance[next.id]; !exists {
					distance[next.id] = distance[current] + 1
					queue = append(queue, next.id)
				}
				if !seen[next.id] {
					if len(nodes) >= limit {
						continue
					}
					memory, findErr := h.memRepo.FindByID(r.Context(), next.id)
					if findErr != nil || memory == nil {
						continue
					}
					seen[next.id] = true
					nodes = append(nodes, GraphNode{ID: memory.ID, MemoryID: memory.ID, Label: graphLabel(memory.Summary, memory.Content, 120), Category: string(memory.Category), Importance: string(memory.Importance), CommunityID: -1, CreatedAt: memory.CreatedAt, RecordedAt: memory.RecordedAt, HopDistance: distance[next.id], Score: next.strength})
				}
				key := current + "→" + next.id
				if !edgeSet[key] {
					edgeSet[key] = true
					edges = append(edges, GraphEdge{Source: current, Target: next.id, Type: next.relType, Strength: next.strength})
				}
			}
		}
	}

	writeJSON(w, http.StatusOK, GraphResponse{
		Nodes:            nodes,
		Edges:            edges,
		Total:            len(nodes),
		TenantID:         tenantID,
		ProjectionStatus: projectionStatus,
		Warnings:         warnings,
	})
}

// Timeline handles GET /v1/graph/timeline?from=RFC3339&to=RFC3339&limit=200
func (h *GraphViewHandler) Timeline(w http.ResponseWriter, r *http.Request) {
	if h.memRepo == nil {
		writeError(w, http.StatusServiceUnavailable, "memory repository not available")
		return
	}
	tenantID := tenant.FromContext(r.Context())

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}

	from := time.Time{}
	to := time.Now()
	if f := r.URL.Query().Get("from"); f != "" {
		if ts, err := time.Parse(time.RFC3339, f); err == nil {
			from = ts
		}
	}
	if t := r.URL.Query().Get("to"); t != "" {
		if ts, err := time.Parse(time.RFC3339, t); err == nil {
			to = ts
		}
	}

	history, err := h.memRepo.FindHistoryRange(r.Context(), from, to, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	nodes, edges := buildTimelineGraph(history)

	writeJSON(w, http.StatusOK, GraphResponse{
		Nodes:    nodes,
		Edges:    edges,
		Total:    len(nodes),
		TenantID: tenantID,
	})
}

func buildTimelineGraph(history []postgres.MemoryHistoryEntry) ([]GraphNode, []GraphEdge) {
	nodes := make([]GraphNode, 0, len(history))
	versionNodes := make(map[string][]GraphNode)
	for _, entry := range history {
		m := entry.Memory
		nodeID := fmt.Sprintf("%s@%d", m.ID, entry.SystemFrom.UnixNano())
		systemFrom := entry.SystemFrom
		nodes = append(nodes, GraphNode{
			ID:              nodeID,
			MemoryID:        m.ID,
			Label:           graphLabel(m.Summary, m.Content, 120),
			Category:        string(m.Category),
			Importance:      string(m.Importance),
			CommunityID:     -1,
			AccessCount:     m.AccessCount,
			HelpfulCount:    m.HelpfulCount,
			NotHelpfulCount: m.NotHelpfulCount,
			EmotionalWeight: m.EmotionalWeight,
			CreatedAt:       m.CreatedAt,
			ValidFrom:       m.ValidFrom,
			ValidTo:         m.ValidTo,
			SupersededBy:    m.SupersededBy,
			Tags:            m.Tags,
			Version:         m.Version,
			Operation:       entry.Operation,
			SystemFrom:      &systemFrom,
			SystemTo:        entry.SystemTo,
			RecordedAt:      systemFrom,
		})
		versionNodes[m.ID] = append(versionNodes[m.ID], nodes[len(nodes)-1])
	}
	edges := make([]GraphEdge, 0)
	for _, versions := range versionNodes {
		for i := 0; i+1 < len(versions); i++ {
			edges = append(edges, GraphEdge{Source: versions[i].ID, Target: versions[i+1].ID, Type: "VERSION", Strength: 1})
		}
	}
	for _, versions := range versionNodes {
		for _, node := range versions {
			if node.SupersededBy == "" || len(versionNodes[node.SupersededBy]) == 0 {
				continue
			}
			targetVersions := versionNodes[node.SupersededBy]
			target := targetVersions[len(targetVersions)-1]
			edges = append(edges, GraphEdge{
				Source: node.ID, Target: target.ID, Type: "SUPERSEDES", Strength: 1.0,
			})
		}
	}

	return nodes, edges
}

func graphLabel(summary, content string, max int) string {
	l := summary
	if strings.TrimSpace(l) == "" {
		l = content
	}
	l = strings.TrimSpace(l)
	if len(l) > max {
		return l[:max-3] + "..."
	}
	return l
}
