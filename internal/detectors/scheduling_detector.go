package detectors

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"

	"github.com/kuberadar/kuberadar/internal/constants"
	"github.com/kuberadar/kuberadar/internal/detectors/utils"
	"github.com/kuberadar/kuberadar/internal/model"
)

// SchedulingDetector detects unschedulable pods.
type SchedulingDetector struct{}

// Name returns the detector name.
func (SchedulingDetector) Name() string {
	return "scheduling"
}

// Target returns the scope type this detector applies to.
func (SchedulingDetector) Target() model.ScopeType {
	return model.ScopePod
}

// Detect runs the detection logic.
func (d SchedulingDetector) Detect(ctx model.Context) []model.Issue {
	if ctx.Pod == nil {
		return nil
	}
	if !utils.IsUnschedulable(ctx.Pod) {
		return nil
	}
	var evidence []model.Evidence
	var recommendation string

	var rootCause string
	if len(ctx.SchedulingErrors) > 0 {
		// Rich evidence from ContextBuilder (diagnose mode) — structured block with hierarchy
		for _, err := range ctx.SchedulingErrors {
			var b strings.Builder
			b.WriteString("Missing dependency\n")
			b.WriteString("─────────────────\n")
			b.WriteString(fmt.Sprintf("PVC: %s\n\n", err.ClaimName))
			b.WriteString("Cluster verification\n")
			b.WriteString("────────────────────\n")
			var verifications []string
			if !err.Exists {
				v := fmt.Sprintf("PVC %s not found in namespace %s", err.ClaimName, ctx.Pod.Namespace)
				verifications = append(verifications, v)
				rootCause = v
			} else {
				if !err.IsBound {
					verifications = append(verifications, fmt.Sprintf("PVC %s exists but is not bound to a PersistentVolume", err.ClaimName))
				}
				if err.StorageClassName != "" && !err.StorageClassExists {
					v := fmt.Sprintf("StorageClass %q not found", err.StorageClassName)
					verifications = append(verifications, v)
					rootCause = v
				} else if !err.IsBound {
					rootCause = fmt.Sprintf("PVC %s exists but is not bound to a PersistentVolume", err.ClaimName)
				}
				if len(verifications) == 0 {
					verifications = append(verifications, fmt.Sprintf("PVC %s exists in namespace %s — check storage class or binding.", err.ClaimName, ctx.Pod.Namespace))
				}
			}
			for _, v := range verifications {
				b.WriteString(v + "\n")
			}
			b.WriteString("\n")
			if len(err.ReferencedBy) > 0 {
				b.WriteString("Reference location\n")
				b.WriteString("─────────────────\n")
				b.WriteString(strings.Join(err.ReferencedBy, "\n\n"))
			}
			evidence = append(evidence, model.EvidenceFromBlock(strings.TrimRight(b.String(), "\n"), rootCause)...)
		}
		recommendation = buildSchedulingRecommendation(ctx.SchedulingErrors[0], ctx.Pod.Namespace)
	} else {
		// Basic evidence (sweep mode)
		if len(ctx.Pod.Spec.NodeSelector) > 0 {
			var pairs []string
			for k, v := range ctx.Pod.Spec.NodeSelector {
				pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
			}
			evidence = append(evidence, model.EvidenceLine(fmt.Sprintf("Node selector: %s", strings.Join(pairs, ", ")), false))
		} else {
			evidence = append(evidence, model.EvidenceLine("Node selector: (none)", false))
		}
		for _, cond := range ctx.Pod.Status.Conditions {
			if cond.Type == v1.PodScheduled && cond.Status == v1.ConditionFalse && cond.Message != "" {
				evidence = append(evidence, model.EvidenceLine(fmt.Sprintf("Scheduler message: %s", cond.Message), false))
				break
			}
		}
		recommendation = "Check node resources, taints/tolerations, affinity rules, and PVC bindings. Review scheduler events."
	}

	resourceName := ctx.Pod.Namespace + "/" + ctx.Pod.Name
	return []model.Issue{{
		ID:             constants.IssueIDUnschedulablePod,
		Severity:       constants.SeverityWarning,
		ResourceKind:   "Pod",
		ResourceName:   resourceName,
		Message:        "Pod is unschedulable",
		Evidence:       evidence,
		Recommendation: recommendation,
	}}
}

func buildSchedulingRecommendation(err model.PVCErrorInfo, ns string) string {
	if err.StorageClassName != "" && !err.StorageClassExists {
		return fmt.Sprintf("StorageClass %q does not exist — create it or use an existing StorageClass in the PVC spec.\n\nkubectl get storageclass", err.StorageClassName)
	}
	if err.Exists && !err.IsBound {
		return fmt.Sprintf("PVC %s exists but is unbound — check StorageClass, provisioner, or create a matching PV.\n\nkubectl describe pvc %s -n %s", err.ClaimName, err.ClaimName, ns)
	}
	if err.Exists {
		return fmt.Sprintf("PVC %s exists — verify storage class and binding.\n\nkubectl describe pvc %s -n %s", err.ClaimName, err.ClaimName, ns)
	}
	return fmt.Sprintf("Create the missing PVC:\n\nkubectl apply -f - <<EOF\napiVersion: v1\nkind: PersistentVolumeClaim\nmetadata:\n  name: %s\n  namespace: %s\nspec:\n  accessModes: [ReadWriteOnce]\n  resources:\n    requests:\n      storage: 1Gi\nEOF\n\nor fix the reference in the pod spec.", err.ClaimName, ns)
}
