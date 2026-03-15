package output

import (
	"os"
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/Vemula-Rohith/kuberadar/internal/model"
)

// Format types.
const (
	FormatTable = "table"
	FormatJSON  = "json"
)

// Options configures output behavior.
type Options struct {
	SinglePod bool // true when diagnosing one specific resource
	Diagnose  bool // true for --diagnose: show full evidence and recommendation
}

// Print writes the diagnosis to stdout using the given format.
// Uses lipgloss for TTY-aware color output (strips ANSI when piping).
func Print(d *model.Diagnosis, format string, opts Options) error {
	switch strings.ToLower(format) {
	case FormatJSON:
		return WriteJSON(os.Stdout, d)
	case FormatTable:
		lipgloss.Println(FormatAsTable(d, opts))
		return nil
	default:
		lipgloss.Println(FormatAsTable(d, opts))
		return nil
	}
}
