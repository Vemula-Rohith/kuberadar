package cli

import (
	"os"

	"github.com/Vemula-Rohith/kuberadar/internal/constants"
	"github.com/Vemula-Rohith/kuberadar/internal/model"
)

// Exit codes for CI integration:
// 0 = no issues
// 1 = warnings only
// 2 = critical issues
const (
	ExitSuccess   = 0
	ExitWarnings  = 1
	ExitCritical  = 2
)

// ExitWithCode exits the process based on diagnosis severity.
func ExitWithCode(d *model.Diagnosis) {
	hasCritical := false
	hasWarning := false
	for _, i := range d.Issues {
		if i.Severity == constants.SeverityCritical {
			hasCritical = true
		}
		if i.Severity == constants.SeverityWarning {
			hasWarning = true
		}
	}
	if hasCritical {
		os.Exit(ExitCritical)
	}
	if hasWarning {
		os.Exit(ExitWarnings)
	}
	os.Exit(ExitSuccess)
}
