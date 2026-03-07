package commands

import (
	"fmt"

	"github.com/integraltech/brainsentry/internal/dto"
	"github.com/spf13/cobra"
)

func newCorrectCmd(a *App) *cobra.Command {
	var (
		reason           string
		correctedContent string
		flaggedBy        string
	)

	cmd := &cobra.Command{
		Use:   "correct [memoryID]",
		Short: "Flag a memory as incorrect",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if a.Corrector == nil {
				return fmt.Errorf("correction service not configured")
			}
			if reason == "" {
				return fmt.Errorf("--reason is required")
			}
			if err := a.validateTenantID(); err != nil {
				return err
			}

			req := dto.FlagMemoryRequest{
				Reason:           reason,
				CorrectedContent: correctedContent,
				FlaggedBy:        flaggedBy,
			}

			ctx, cancel := a.newContext()
			defer cancel()

			correction, err := a.Corrector.FlagMemory(ctx, args[0], req)
			if err != nil {
				return fmt.Errorf("flagging memory: %w", err)
			}

			w := cmd.OutOrStdout()
			if a.Output == "json" {
				return printJSON(w, correction)
			}
			fmt.Fprintf(w, "Flagged memory %s (correction: %s)\n", args[0], correction.ID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&reason, "reason", "r", "", "Reason for flagging (required)")
	cmd.Flags().StringVar(&correctedContent, "corrected-content", "", "Corrected content")
	cmd.Flags().StringVar(&flaggedBy, "flagged-by", "cli", "Flagged by")

	return cmd
}
