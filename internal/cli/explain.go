package cli

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/Vemula-Rohith/kuberadar/internal/explain"
	"github.com/Vemula-Rohith/kuberadar/internal/output"
)

var explainCmd = &cobra.Command{
	Use:   "explain [issue-id]",
	Short: "Explain an issue ID and how to resolve it",
	Long:  "Explain provides documentation for KubeRadar issue IDs (e.g. KR001, KR003).",
	RunE:  runExplain,
}

func init() {
	rootCmd.AddCommand(explainCmd)
}

func runExplain(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		lipgloss.Println("Usage: kuberadar explain <issue-id>")
		lipgloss.Println("\n" + output.IconResource + " Available issue IDs:")
		for id, entry := range explain.Registry {
			lipgloss.Printf("  %s — %s\n", id, entry.Name)
		}
		return nil
	}

	id := strings.ToUpper(args[0])
	entry, ok := explain.Registry[id]
	if !ok {
		return fmt.Errorf("unknown issue ID: %s (use 'kuberadar explain' to list available IDs)", id)
	}

	lipgloss.Printf("%s %s — %s\n\n", output.IconSweep, entry.ID, entry.Name)
	lipgloss.Println(output.IconEvidence + " Description:")
	lipgloss.Println("  " + entry.Description)
	lipgloss.Println("\n" + output.IconSummary + " Common causes:")
	for _, c := range entry.CommonCauses {
		lipgloss.Printf("  • %s\n", c)
	}
	lipgloss.Println("\n" + output.IconRecommend + " Recommended actions:")
	for _, a := range entry.RecommendedActions {
		lipgloss.Printf("  • %s\n", a)
	}
	return nil
}
