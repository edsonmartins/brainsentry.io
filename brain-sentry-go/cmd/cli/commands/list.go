package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newListCmd(a *App) *cobra.Command {
	var (
		page int
		size int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List memories",
		RunE: func(cmd *cobra.Command, args []string) error {
			if a.Lister == nil {
				return fmt.Errorf("list service not configured")
			}
			if err := a.validateTenantID(); err != nil {
				return err
			}

			ctx, cancel := a.newContext()
			defer cancel()

			resp, err := a.Lister.ListMemories(ctx, page, size)
			if err != nil {
				return fmt.Errorf("listing: %w", err)
			}

			w := cmd.OutOrStdout()
			if a.Output == "json" {
				return printJSON(w, resp)
			}

			headers := []string{"ID", "Summary", "Category", "Importance"}
			rows := make([][]string, len(resp.Memories))
			for i, m := range resp.Memories {
				summary := m.Summary
				if summary == "" {
					summary = truncate(m.Content, 50)
				}
				rows[i] = []string{
					truncate(m.ID, 12),
					truncate(summary, 50),
					string(m.Category),
					string(m.Importance),
				}
			}
			printTable(w, headers, rows)
			fmt.Fprintf(w, "\nPage %d/%d (%d total)\n", resp.Page+1, resp.TotalPages, resp.TotalElements)
			return nil
		},
	}

	cmd.Flags().IntVarP(&page, "page", "p", 0, "Page number (0-based)")
	cmd.Flags().IntVarP(&size, "size", "s", 20, "Page size")

	return cmd
}
