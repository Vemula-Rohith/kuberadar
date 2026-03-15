package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/kuberadar/kuberadar/internal/model"
	"github.com/kuberadar/kuberadar/internal/output"
)

var sweepCmd = &cobra.Command{
	Use:   "sweep",
	Short: "Sweep all pods in namespace for issues",
	RunE:  runSweep,
}

func init() {
	rootCmd.AddCommand(sweepCmd)
}

func runSweep(cmd *cobra.Command, _ []string) error {
	scope := model.Scope{
		Type:      model.ScopePod,
		Namespace: namespace,
	}
	diagnosis, err := app.engine.Run(context.Background(), scope)
	if err != nil {
		return err
	}
	if err := output.Print(diagnosis, outputFmt, output.Options{}); err != nil {
		return err
	}
	ExitWithCode(diagnosis)
	return nil
}
