package commands

import (
	"context"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
)

// MemoryCreator creates memories.
type MemoryCreator interface {
	CreateMemory(ctx context.Context, req dto.CreateMemoryRequest) (*domain.Memory, error)
}

// MemorySearcher searches memories.
type MemorySearcher interface {
	SearchMemories(ctx context.Context, req dto.SearchRequest) (*dto.SearchResponse, error)
}

// MemoryLister lists memories with pagination.
type MemoryLister interface {
	ListMemories(ctx context.Context, page, size int) (*dto.MemoryListResponse, error)
}

// MemoryUpdater updates existing memories.
type MemoryUpdater interface {
	UpdateMemory(ctx context.Context, id string, req dto.UpdateMemoryRequest) (*domain.Memory, error)
}

// MemoryCorrector handles flagging and reviewing memory corrections.
type MemoryCorrector interface {
	FlagMemory(ctx context.Context, memoryID string, req dto.FlagMemoryRequest) (*domain.MemoryCorrection, error)
	ReviewCorrection(ctx context.Context, memoryID string, req dto.ReviewCorrectionRequest) (*domain.Memory, error)
}
