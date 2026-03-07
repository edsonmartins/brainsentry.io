package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/integraltech/brainsentry/internal/service"
)

// BenchmarkHandler handles benchmark endpoints.
type BenchmarkHandler struct {
	benchmarkService *service.BenchmarkService
	memoryService    *service.MemoryService
}

// NewBenchmarkHandler creates a new BenchmarkHandler.
func NewBenchmarkHandler(benchmarkService *service.BenchmarkService, memoryService *service.MemoryService) *BenchmarkHandler {
	return &BenchmarkHandler{
		benchmarkService: benchmarkService,
		memoryService:    memoryService,
	}
}

// RunBenchmark handles POST /v1/benchmark/run — evaluates retrieval quality with a synthetic dataset.
func (h *BenchmarkHandler) RunBenchmark(w http.ResponseWriter, r *http.Request) {
	var req struct {
		QueryCount       int      `json:"queryCount"`
		ExpectedPerQuery int      `json:"expectedPerQuery"`
		K                int      `json:"k"`
		Categories       []string `json:"categories"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req.QueryCount = 10
		req.ExpectedPerQuery = 5
		req.K = 10
	}

	if req.QueryCount <= 0 {
		req.QueryCount = 10
	}
	if req.ExpectedPerQuery <= 0 {
		req.ExpectedPerQuery = 5
	}
	if req.K <= 0 {
		req.K = 10
	}
	if len(req.Categories) == 0 {
		req.Categories = []string{"INSIGHT", "WARNING", "KNOWLEDGE", "ACTION", "CONTEXT", "REFERENCE", "GENERAL"}
	}

	dataset := service.GenerateSyntheticDataset("api-benchmark", req.QueryCount, req.ExpectedPerQuery, req.Categories)

	searchFn := func(query string, k int) ([]string, time.Duration, error) {
		start := time.Now()
		resp, err := h.memoryService.SearchMemories(r.Context(), dto.SearchRequest{
			Query: query,
			Limit: k,
		})
		elapsed := time.Since(start)
		if err != nil {
			return nil, elapsed, err
		}
		ids := make([]string, len(resp.Results))
		for i, m := range resp.Results {
			ids[i] = m.ID
		}
		return ids, elapsed, nil
	}

	report, err := h.benchmarkService.RunBenchmark(dataset, searchFn, req.K)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "benchmark failed: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, report)
}
