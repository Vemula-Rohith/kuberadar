package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Vemula-Rohith/kuberadar/internal/model"
	"github.com/Vemula-Rohith/kuberadar/internal/output"
)

var (
	diagnose bool
)

var podCmd = &cobra.Command{
	Use:   "pod [name]",
	Short: "Diagnose a specific pod or all pods in namespace",
	RunE:  runPod,
}

func init() {
	rootCmd.AddCommand(podCmd)
	podCmd.Flags().BoolVar(&diagnose, "diagnose", false, "Show full evidence and recommendation for the pod")
}

func runPod(cmd *cobra.Command, args []string) error {
	scope := model.Scope{
		Type:      model.ScopePod,
		Namespace: namespace,
		Diagnose:  diagnose,
	}
	if len(args) > 0 {
		scope.Name = args[0]
	}
	diagnosis, err := app.engine.Run(context.Background(), scope)
	if err != nil {
		return fmt.Errorf("diagnosis failed: %w", err)
	}
	opts := output.Options{SinglePod: scope.Name != "", Diagnose: diagnose}
	if err := output.Print(diagnosis, outputFmt, opts); err != nil {
		return err
	}
	ExitWithCode(diagnosis)
	return nil
}
