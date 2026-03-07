package commands

import (
	"fmt"

	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/spf13/cobra"
)

func newReviewCmd(a *App) *cobra.Command {
	var (
		action   string
		notes    string
		reviewer string
	)

	cmd := &cobra.Command{
		Use:   "review [memoryID]",
		Short: "Review a flagged memory correction",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if a.Corrector == nil {
				return fmt.Errorf("correction service not configured")
			}
			if action != "approve" && action != "reject" {
				return fmt.Errorf("--action must be 'approve' or 'reject'")
			}
			if err := a.validateTenantID(); err != nil {
				return err
			}

			req := dto.ReviewCorrectionRequest{
				Action:      action,
				ReviewNotes: notes,
				ReviewedBy:  reviewer,
			}

			ctx, cancel := a.newContext()
			defer cancel()

			mem, err := a.Corrector.ReviewCorrection(ctx, args[0], req)
			if err != nil {
				return fmt.Errorf("reviewing correction: %w", err)
			}

			w := cmd.OutOrStdout()
			if a.Output == "json" {
				return printJSON(w, mem)
			}
			fmt.Fprintf(w, "Reviewed memory %s: %s\n", mem.ID, action)
			return nil
		},
	}

	cmd.Flags().StringVarP(&action, "action", "a", "", "Action: approve or reject (required)")
	cmd.Flags().StringVar(&notes, "notes", "", "Review notes")
	cmd.Flags().StringVar(&reviewer, "reviewer", "cli", "Reviewer identity")

	return cmd
}
