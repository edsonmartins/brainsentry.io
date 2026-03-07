package commands

import (
	"fmt"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/spf13/cobra"
)

func newSearchCmd(a *App) *cobra.Command {
	var (
		limit         int
		categories    []string
		tags          []string
		minImportance string
	)

	cmd := &cobra.Command{
		Use:   "search [query]",
		Short: "Search memories",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if a.Searcher == nil {
				return fmt.Errorf("search service not configured")
			}
			if err := a.validateTenantID(); err != nil {
				return err
			}

			cats := make([]domain.MemoryCategory, len(categories))
			for i, c := range categories {
				cats[i] = domain.MemoryCategory(c)
			}

			req := dto.SearchRequest{
				Query:         args[0],
				Limit:         limit,
				Categories:    cats,
				Tags:          tags,
				MinImportance: domain.ImportanceLevel(minImportance),
			}

			ctx, cancel := a.newContext()
			defer cancel()

			resp, err := a.Searcher.SearchMemories(ctx, req)
			if err != nil {
				return fmt.Errorf("searching: %w", err)
			}

			w := cmd.OutOrStdout()
			if a.Output == "json" {
				return printJSON(w, resp)
			}

			headers := []string{"ID", "Summary", "Category", "Importance"}
			rows := make([][]string, len(resp.Results))
			for i, r := range resp.Results {
				summary := r.Summary
				if summary == "" {
					summary = truncate(r.Content, 50)
				}
				rows[i] = []string{
					truncate(r.ID, 12),
					truncate(summary, 50),
					string(r.Category),
					string(r.Importance),
				}
			}
			printTable(w, headers, rows)
			fmt.Fprintf(w, "\n%d results (%dms)\n", resp.Total, resp.SearchTimeMs)
			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 10, "Max results")
	cmd.Flags().StringSliceVar(&categories, "categories", nil, "Filter by categories")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Filter by tags")
	cmd.Flags().StringVar(&minImportance, "min-importance", "", "Minimum importance level")

	return cmd
}
