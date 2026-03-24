package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Vemula-Rohith/kuberadar/internal/model"
	"github.com/Vemula-Rohith/kuberadar/internal/output"
	"github.com/Vemula-Rohith/kuberadar/internal/state"
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
	diagnosis, err := engineRun(cmd, scope)
	if err != nil {
		return err
	}
	if err := state.WriteLastSweep(state.EntriesFromDiagnosis(diagnosis)); err != nil {
		// Non-fatal: sweep output still useful without index file
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "warning: could not save sweep index: %v\n", err)
	}
	if err := output.Print(diagnosis, outputFmt, output.Options{}); err != nil {
		return err
	}
	FinishDiagnosis(diagnosis)
	return nil
}
