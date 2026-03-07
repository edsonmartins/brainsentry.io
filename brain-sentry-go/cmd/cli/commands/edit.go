package commands

import (
	"fmt"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/spf13/cobra"
)

func newEditCmd(a *App) *cobra.Command {
	var (
		content    string
		summary    string
		category   string
		importance string
		tags       []string
		reason     string
	)

	cmd := &cobra.Command{
		Use:   "edit [memoryID]",
		Short: "Edit an existing memory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if a.Updater == nil {
				return fmt.Errorf("update service not configured")
			}
			if err := a.validateTenantID(); err != nil {
				return err
			}

			req := dto.UpdateMemoryRequest{
				Content:      content,
				Summary:      summary,
				Category:     domain.MemoryCategory(category),
				Importance:   domain.ImportanceLevel(importance),
				Tags:         tags,
				ChangeReason: reason,
			}

			ctx, cancel := a.newContext()
			defer cancel()

			mem, err := a.Updater.UpdateMemory(ctx, args[0], req)
			if err != nil {
				return fmt.Errorf("updating memory: %w", err)
			}

			w := cmd.OutOrStdout()
			if a.Output == "json" {
				return printJSON(w, mem)
			}
			fmt.Fprintf(w, "Updated memory: %s\n", mem.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&content, "content", "", "New content")
	cmd.Flags().StringVar(&summary, "summary", "", "New summary")
	cmd.Flags().StringVarP(&category, "category", "c", "", "New category")
	cmd.Flags().StringVarP(&importance, "importance", "i", "", "New importance")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "New tags")
	cmd.Flags().StringVar(&reason, "reason", "", "Change reason")

	return cmd
}
