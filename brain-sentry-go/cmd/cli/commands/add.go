package commands

import (
	"fmt"

	"github.com/integraltech/brainsentry/internal/domain"
	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/spf13/cobra"
)

func newAddCmd(a *App) *cobra.Command {
	var (
		category   string
		importance string
		memType    string
		tags       []string
		source     string
		summary    string
		code       string
		lang       string
	)

	cmd := &cobra.Command{
		Use:   "add [content]",
		Short: "Create a new memory",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if a.Creator == nil {
				return fmt.Errorf("memory service not configured")
			}
			if err := a.validateTenantID(); err != nil {
				return err
			}

			req := dto.CreateMemoryRequest{
				Content:             args[0],
				Summary:             summary,
				Category:            domain.MemoryCategory(category),
				Importance:          domain.ImportanceLevel(importance),
				MemoryType:          domain.MemoryType(memType),
				Tags:                tags,
				SourceType:          source,
				CodeExample:         code,
				ProgrammingLanguage: lang,
			}

			ctx, cancel := a.newContext()
			defer cancel()

			mem, err := a.Creator.CreateMemory(ctx, req)
			if err != nil {
				return fmt.Errorf("creating memory: %w", err)
			}

			w := cmd.OutOrStdout()
			if a.Output == "json" {
				return printJSON(w, mem)
			}
			fmt.Fprintf(w, "Created memory: %s\n", mem.ID)
			fmt.Fprintf(w, "Category: %s | Importance: %s\n", mem.Category, mem.Importance)
			return nil
		},
	}

	cmd.Flags().StringVarP(&category, "category", "c", "", "Memory category (KNOWLEDGE, INSIGHT, DECISION, etc.)")
	cmd.Flags().StringVarP(&importance, "importance", "i", "", "Importance level (CRITICAL, IMPORTANT, MINOR)")
	cmd.Flags().StringVarP(&memType, "type", "t", "", "Memory type (SEMANTIC, EPISODIC, PROCEDURAL, etc.)")
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Tags (comma-separated)")
	cmd.Flags().StringVar(&source, "source", "cli", "Source type")
	cmd.Flags().StringVar(&summary, "summary", "", "Summary")
	cmd.Flags().StringVar(&code, "code", "", "Code example")
	cmd.Flags().StringVar(&lang, "lang", "", "Programming language")

	return cmd
}
