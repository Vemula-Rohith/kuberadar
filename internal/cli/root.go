package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Vemula-Rohith/kuberadar/internal/engine"
	"github.com/Vemula-Rohith/kuberadar/internal/output"
	"github.com/Vemula-Rohith/kuberadar/internal/providers"
)

var (
	kubeconfig string
	namespace  string
	outputFmt  string
)

// app holds shared state for CLI commands.
var app struct {
	engine *engine.Engine
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "kuberadar",
	Short: "Kubernetes debugging and diagnostics CLI",
	Long:  "KubeRadar helps developers diagnose and debug issues in Kubernetes clusters.",
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		if cmd.Name() == "version" || cmd.Name() == "explain" {
			return nil
		}
		ns := namespace
		if ns == "" {
			ns = providers.GetDefaultNamespace(kubeconfig)
		}
		client, _, err := providers.NewClient(providers.Config{Kubeconfig: kubeconfig})
		if err != nil {
			return fmt.Errorf("failed to create kube client: %w", err)
		}
		podProvider := providers.NewPodProvider(client)
		deployProvider := providers.NewDeploymentProvider(client)
		eventProvider := providers.NewEventProvider(client)
		nodeProvider := providers.NewNodeProvider(client)
		configProvider := providers.NewConfigProvider(client)
		pvcProvider := providers.NewPVCProvider(client)
		storageClassProvider := providers.NewStorageClassProvider(client)
		cb := engine.NewContextBuilder(eventProvider, nodeProvider, podProvider, configProvider, pvcProvider, storageClassProvider)
		app.engine = engine.NewEngine(podProvider, deployProvider, cb)
		namespace = ns
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "", "Path to kubeconfig file")
	rootCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace (default from kubeconfig context)")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", output.FormatTable, "Output format: table, json")
}
