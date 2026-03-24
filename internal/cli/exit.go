package cli

import (
	"os"

	"github.com/Vemula-Rohith/kuberadar/internal/constants"
	"github.com/Vemula-Rohith/kuberadar/internal/model"
)

// Exit codes when --fail-on-issues is set (CI):
// 0 = no issues
// 1 = warnings only
// 2 = critical issues
const (
	ExitSuccess  = 0
	ExitWarnings = 1
	ExitCritical = 2
)

// FinishDiagnosis exits the process if --fail-on-issues is set and severity demands it.
// Otherwise the command returns normally (exit 0) — success means the scan completed,
// not “no issues found”.
func FinishDiagnosis(d *model.Diagnosis) {
	if !failOnIssues {
		return
	}
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
