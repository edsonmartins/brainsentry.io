package commands

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/integraltech/brainsentry/pkg/tenant"
)

const (
	defaultTimeout = 30 * time.Second
	maxImportFileSize = 100 * 1024 * 1024 // 100 MB
)

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// App holds CLI dependencies for all subcommands.
type App struct {
	Creator   MemoryCreator
	Searcher  MemorySearcher
	Lister    MemoryLister
	Updater   MemoryUpdater
	Corrector MemoryCorrector
	TenantID  string
	Output    string // "table", "json", or "plain"
	Writer    io.Writer
	Timeout   time.Duration
}

// newContext creates a context with tenant and timeout.
func (a *App) newContext() (context.Context, context.CancelFunc) {
	timeout := a.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	ctx = tenant.WithTenant(ctx, a.TenantID)
	return ctx, cancel
}

// validateTenantID checks that the tenant ID is a valid UUID.
func (a *App) validateTenantID() error {
	if !uuidRegex.MatchString(a.TenantID) {
		return fmt.Errorf("invalid tenant ID %q: must be a valid UUID", a.TenantID)
	}
	return nil
}
