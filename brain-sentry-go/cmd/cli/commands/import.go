package commands

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/spf13/cobra"
)

// importFile represents the JSON structure for bulk import.
type importFile struct {
	Memories []dto.CreateMemoryRequest `json:"memories"`
}

func newImportCmd(a *App) *cobra.Command {
	var (
		skipDuplicates bool
		dryRun         bool
	)

	cmd := &cobra.Command{
		Use:   "import [file.json]",
		Short: "Bulk import memories from JSON file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if a.Creator == nil {
				return fmt.Errorf("memory service not configured")
			}
			if err := a.validateTenantID(); err != nil {
				return err
			}

			// Check file size before reading
			info, err := os.Stat(args[0])
			if err != nil {
				return fmt.Errorf("reading file: %w", err)
			}
			if info.Size() > maxImportFileSize {
				return fmt.Errorf("file too large: %d bytes (max %d bytes / 100MB)", info.Size(), maxImportFileSize)
			}

			data, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("reading file: %w", err)
			}

			var file importFile
			if err := json.Unmarshal(data, &file); err != nil {
				return fmt.Errorf("parsing JSON: %w", err)
			}

			if len(file.Memories) == 0 {
				return fmt.Errorf("no memories found in file")
			}

			w := cmd.OutOrStdout()

			if dryRun {
				fmt.Fprintf(w, "Dry run: would import %d memories\n", len(file.Memories))
				return nil
			}

			ctx, cancel := a.newContext()
			defer cancel()

			var imported, failed int
			for _, req := range file.Memories {
				_, err := a.Creator.CreateMemory(ctx, req)
				if err != nil {
					if skipDuplicates {
						failed++
						continue
					}
					return fmt.Errorf("importing memory: %w", err)
				}
				imported++
			}

			fmt.Fprintf(w, "Imported: %d, Skipped: %d\n", imported, failed)
			return nil
		},
	}

	cmd.Flags().BoolVar(&skipDuplicates, "skip-duplicates", false, "Skip duplicate memories")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without importing")

	return cmd
}
