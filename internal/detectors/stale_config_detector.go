package detectors

import (
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"

	"github.com/Vemula-Rohith/kuberadar/internal/constants"
	"github.com/Vemula-Rohith/kuberadar/internal/detectors/utils"
	"github.com/Vemula-Rohith/kuberadar/internal/model"
)

const staleConfigMinSkew = time.Minute

// StaleConfigDetector flags ConfigMaps/Secrets whose content likely changed after the
// pod's effective start time (mounted/env data may be stale until restart).
type StaleConfigDetector struct{}

func (StaleConfigDetector) Name() string { return "StaleConfig" }

func (StaleConfigDetector) Target() model.ScopeType { return model.ScopePod }

func (StaleConfigDetector) Detect(ctx model.Context) []model.Issue {
	pod := ctx.Pod
	if pod == nil || ctx.ConfigMaps == nil && ctx.Secrets == nil {
		return nil
	}
	podStart := pod.CreationTimestamp.Time
	if pod.Status.StartTime != nil && !pod.Status.StartTime.IsZero() {
		podStart = pod.Status.StartTime.Time
	}
	if podStart.IsZero() {
		return nil
	}
	var issues []model.Issue
	if ctx.ConfigMaps != nil {
		for _, name := range utils.GetConfigMapRefs(pod) {
			cm, ok := ctx.ConfigMaps[name]
			if !ok {
				continue
			}
			if issue := staleConfigIssue(pod, "ConfigMap", name, utils.ApproximateConfigResourceUpdateTime(cm.ObjectMeta), podStart, cm.ResourceVersion); issue != nil {
				issues = append(issues, *issue)
			}
		}
	}
	if ctx.Secrets != nil {
		for _, name := range utils.GetSecretRefs(pod) {
			sec, ok := ctx.Secrets[name]
			if !ok {
				continue
			}
			if issue := staleConfigIssue(pod, "Secret", name, utils.ApproximateConfigResourceUpdateTime(sec.ObjectMeta), podStart, sec.ResourceVersion); issue != nil {
				issues = append(issues, *issue)
			}
		}
	}
	return issues
}

func staleConfigIssue(pod *v1.Pod, kind, name string, resourceUpdated, podStart time.Time, resourceVersion string) *model.Issue {
	if resourceUpdated.IsZero() || !resourceUpdated.After(podStart) {
		return nil
	}
	if resourceUpdated.Sub(podStart) < staleConfigMinSkew {
		return nil
	}
	ref := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
	return &model.Issue{
		ID:           constants.IssueIDStaleConfig,
		Severity:     constants.SeverityWarning,
		ResourceKind: "Pod",
		ResourceName: ref,
		Message:      "Stale configuration detected",
		LikelyCause:  "Configuration was updated after this pod started; the workload may still be using older data until the pod is recreated.",
		Evidence: []model.Evidence{
			model.EvidenceLine(fmt.Sprintf("%s: %s", kind, name), false),
			model.EvidenceLine(fmt.Sprintf("Pod started: %s", podStart.UTC().Format(time.RFC3339)), false),
			model.EvidenceLine(fmt.Sprintf("Resource updated (approx): %s", resourceUpdated.UTC().Format(time.RFC3339)), false),
			model.EvidenceLine(fmt.Sprintf("ResourceVersion: %s", resourceVersion), false),
		},
		Recommendation: "Restart the pod (or roll the workload) so it picks up the latest configuration.",
	}
}
