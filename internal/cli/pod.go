package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Vemula-Rohith/kuberadar/internal/model"
	"github.com/Vemula-Rohith/kuberadar/internal/output"
	"github.com/Vemula-Rohith/kuberadar/internal/state"
)

var (
	diagnose bool
)

var podCmd = &cobra.Command{
	Use:   "pod [name|index]",
	Short: "Diagnose a pod by name or by sweep row number (1, 2, … after kuberadar sweep)",
	Long: `Diagnose pods in the current namespace, or a single pod by name.

After kuberadar sweep, each issue row is numbered. Use that number to deep-dive without
retyping the pod name, e.g. kuberadar pod 1 --diagnose

Namespace for a numeric index comes from the saved sweep (not from -n).`,
	RunE: runPod,
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
		arg := args[0]
		if state.IsSweepIndexSyntax(arg) {
			idx, err := state.ParseSweepIndex(arg)
			if err != nil {
				return fmt.Errorf("invalid sweep index %q", arg)
			}
			if ns, podName, err := state.ResolveSweepIndex(idx); err == nil {
				scope.Namespace = ns
				scope.Name = podName
			} else {
				// No sweep / out of range: treat as a pod whose name is digits (e.g. "1")
				scope.Name = arg
			}
		} else {
			scope.Name = arg
		}
	}
	diagnosis, err := engineRun(cmd, scope)
	if err != nil {
		return fmt.Errorf("diagnosis failed: %w", err)
	}
	opts := output.Options{SinglePod: scope.Name != "", Diagnose: diagnose}
	if err := output.Print(diagnosis, outputFmt, opts); err != nil {
		return err
	}
	FinishDiagnosis(diagnosis)
	return nil
}
