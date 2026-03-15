package detectors

import (
	"fmt"

	v1 "k8s.io/api/core/v1"

	"github.com/Vemula-Rohith/kuberadar/internal/constants"
	"github.com/Vemula-Rohith/kuberadar/internal/detectors/utils"
	"github.com/Vemula-Rohith/kuberadar/internal/model"
)

// OOMKillDetector detects OOMKilled containers.
type OOMKillDetector struct{}

// Name returns the detector name.
func (OOMKillDetector) Name() string {
	return "oomkill"
}

// Target returns the scope type this detector applies to.
func (OOMKillDetector) Target() model.ScopeType {
	return model.ScopePod
}

// Detect runs the detection logic.
func (d OOMKillDetector) Detect(ctx model.Context) []model.Issue {
	if ctx.Pod == nil {
		return nil
	}
	containerSpecs := make(map[string]v1.Container)
	for _, c := range ctx.Pod.Spec.Containers {
		containerSpecs[c.Name] = c
	}
	var evidence []model.Evidence
	var foundOOM bool
	var firstMemLimit bool
	for _, cs := range ctx.Pod.Status.ContainerStatuses {
		oom := (cs.LastTerminationState.Terminated != nil && cs.LastTerminationState.Terminated.Reason == "OOMKilled") ||
			(cs.State.Terminated != nil && cs.State.Terminated.Reason == "OOMKilled")
		if !oom {
			continue
		}
		foundOOM = true
		spec := containerSpecs[cs.Name]
		evidence = append(evidence, model.EvidenceLine(fmt.Sprintf("Container: %s", cs.Name), false))
		memLimit := ""
		memReq := ""
		if q, ok := spec.Resources.Limits[v1.ResourceMemory]; ok {
			memLimit = q.String()
		}
		if q, ok := spec.Resources.Requests[v1.ResourceMemory]; ok {
			memReq = q.String()
		}
		memLimitLine := fmt.Sprintf("Memory limit: %s", utils.OrDefault(memLimit, "—"))
		evidence = append(evidence, model.EvidenceLine(memLimitLine, !firstMemLimit))
		firstMemLimit = true
		evidence = append(evidence, model.EvidenceLine(fmt.Sprintf("Memory request: %s", utils.OrDefault(memReq, "—")), false))
	}
	if foundOOM && ctx.Pod.Spec.NodeName != "" {
		evidence = append([]model.Evidence{model.EvidenceLine(fmt.Sprintf("Node: %s", ctx.Pod.Spec.NodeName), false)}, evidence...)
	}
	// Node memory pressure (pod was scheduled, so we may have node info)
	if ctx.Node != nil {
		for _, cond := range ctx.Node.Status.Conditions {
			if cond.Type == v1.NodeMemoryPressure && cond.Status == v1.ConditionTrue {
				evidence = append(evidence, model.EvidenceLine(fmt.Sprintf("Node memory pressure: %s on %s", cond.Reason, ctx.Node.Name), false))
				break
			}
		}
	}
	if len(evidence) == 0 {
		return nil
	}
	resourceName := ctx.Pod.Namespace + "/" + ctx.Pod.Name
	return []model.Issue{{
		ID:             constants.IssueIDOOMKilled,
		Severity:       constants.SeverityCritical,
		ResourceKind:   "Pod",
		ResourceName:   resourceName,
		Message:        "OOMKilled detected",
		Evidence:       evidence,
		Recommendation: "Increase memory limits or requests for the container. Consider optimizing application memory usage.",
	}}
}
