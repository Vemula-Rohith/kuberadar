package detectors

import (
	"fmt"

	"github.com/kuberadar/kuberadar/internal/constants"
	"github.com/kuberadar/kuberadar/internal/detectors/utils"
	"github.com/kuberadar/kuberadar/internal/model"
)

// CrashLoopDetector detects CrashLoopBackOff pods.
type CrashLoopDetector struct{}

// Name returns the detector name.
func (CrashLoopDetector) Name() string {
	return "crashloop"
}

// Target returns the scope type this detector applies to.
func (CrashLoopDetector) Target() model.ScopeType {
	return model.ScopePod
}

// Detect runs the detection logic.
func (d CrashLoopDetector) Detect(ctx model.Context) []model.Issue {
	if ctx.Pod == nil {
		return nil
	}
	containerName, restartCount, ok := utils.CrashLoopContainer(ctx.Pod)
	if !ok {
		return nil
	}
	resourceName := ctx.Pod.Namespace + "/" + ctx.Pod.Name

	// Explicit labels for multi-container pods
	evidence := []model.Evidence{
		model.EvidenceLine(fmt.Sprintf("Container: %s", containerName), false),
		model.EvidenceLine(fmt.Sprintf("Restart count: %d", restartCount), false),
	}

	// Likely cause when we know it (e.g. OOMKilled)
	var likelyCause string
	if info, ok := utils.GetLastTerminationInfo(ctx.Pod, containerName); ok {
		if info.Reason == "OOMKilled" {
			likelyCause = "Container repeatedly killed due to OOMKilled."
		}
		exitLine := fmt.Sprintf("Exit code: %d", info.ExitCode)
		reasonLine := fmt.Sprintf("Reason: %s", info.Reason)
		if info.Reason != "" {
			evidence = append(evidence, model.EvidenceLine(exitLine, false))
			evidence = append(evidence, model.EvidenceLine(reasonLine, true))
		} else {
			evidence = append(evidence, model.EvidenceLine(exitLine, true))
		}
		if info.Message != "" {
			evidence = append(evidence, model.EvidenceLine(fmt.Sprintf("Message: %s", info.Message), false))
		}
		if !info.FinishedAt.IsZero() {
			evidence = append(evidence, model.EvidenceLine(fmt.Sprintf("Last crash: %s", info.FinishedAt.UTC().Format("2006-01-02T15:04:05Z07:00")), false))
		}
	}

	// Node info (pod is scheduled when crashing)
	if ctx.Pod.Spec.NodeName != "" {
		evidence = append(evidence, model.EvidenceLine(fmt.Sprintf("Node: %s", ctx.Pod.Spec.NodeName), false))
	}

	// Logs (previous or current fallback) - only in diagnose mode
	if ctx.ContainerLogs != nil {
		if cl, ok := ctx.ContainerLogs[containerName]; ok && cl.Content != "" {
			header := "Previous container logs (last 20 lines)"
			if !cl.FromPrevious {
				header = "Current container logs (last 20 lines)"
			}
			evidence = append(evidence, model.EvidenceLine(header+"\n--------------------------------------\n"+cl.Content, false))
		} else {
			evidence = append(evidence, model.EvidenceLine("Previous logs: unavailable", false))
		}
	}

	return []model.Issue{{
		ID:             constants.IssueIDCrashLoopBackOff,
		Severity:       constants.SeverityCritical,
		ResourceKind:   "Pod",
		ResourceName:   resourceName,
		Message:        "CrashLoopBackOff detected",
		LikelyCause:    likelyCause,
		Evidence:       evidence,
		Recommendation: fmt.Sprintf("Fix the root cause from the logs above. Full logs: kubectl logs %s -n %s --previous", ctx.Pod.Name, ctx.Pod.Namespace),
	}}
}
