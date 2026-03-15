package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Vemula-Rohith/kuberadar/internal/model"
	"github.com/Vemula-Rohith/kuberadar/internal/output"
)

var deploymentCmd = &cobra.Command{
	Use:   "deployment [name]",
	Short: "Diagnose a deployment by checking its pods",
	RunE:  runDeployment,
}

func init() {
	rootCmd.AddCommand(deploymentCmd)
	deploymentCmd.Flags().BoolVar(&diagnose, "diagnose", false, "Show full evidence and recommendation")
}

func runDeployment(cmd *cobra.Command, args []string) error {
	scope := model.Scope{
		Type:      model.ScopeDeployment,
		Namespace: namespace,
	}
	if len(args) > 0 {
		scope.Name = args[0]
	} else {
		// Without name, list all pods in namespace (same as sweep)
		scope.Type = model.ScopePod
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
