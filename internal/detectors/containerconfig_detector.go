package detectors

import (
	"fmt"
	"strings"

	"github.com/Vemula-Rohith/kuberadar/internal/constants"
	"github.com/Vemula-Rohith/kuberadar/internal/detectors/utils"
	"github.com/Vemula-Rohith/kuberadar/internal/model"
)

// ContainerConfigDetector detects CreateContainerConfigError.
type ContainerConfigDetector struct{}

// Name returns the detector name.
func (ContainerConfigDetector) Name() string {
	return "containerconfig"
}

// Target returns the scope type this detector applies to.
func (ContainerConfigDetector) Target() model.ScopeType {
	return model.ScopePod
}

// Detect runs the detection logic.
func (d ContainerConfigDetector) Detect(ctx model.Context) []model.Issue {
	if ctx.Pod == nil {
		return nil
	}
	if !utils.IsCreateContainerConfigError(ctx.Pod) {
		return nil
	}
	resourceName := ctx.Pod.Namespace + "/" + ctx.Pod.Name
	var evidence []model.Evidence
	var recommendation string

	var rootCause string
	if len(ctx.ContainerConfigErrors) > 0 {
		// Rich evidence from ContextBuilder (diagnose mode) — single formatted block with hierarchy
		for _, err := range ctx.ContainerConfigErrors {
			var b strings.Builder
			b.WriteString("Missing dependency\n")
			b.WriteString("─────────────────\n")
			b.WriteString(fmt.Sprintf("%s: %s\n\n", err.ResourceType, err.ResourceName))
			b.WriteString("Cluster verification\n")
			b.WriteString("────────────────────\n")
			if err.Exists {
				b.WriteString(fmt.Sprintf("%s %s exists in namespace %s — check key references or RBAC.\n\n", err.ResourceType, err.ResourceName, ctx.Pod.Namespace))
			} else {
				rootCause = fmt.Sprintf("%s %s not found in namespace %s", err.ResourceType, err.ResourceName, ctx.Pod.Namespace)
				b.WriteString(rootCause + "\n\n")
			}
			if len(err.ReferencedBy) > 0 {
				b.WriteString("Reference location\n")
				b.WriteString("─────────────────\n")
				b.WriteString(strings.Join(err.ReferencedBy, "\n\n"))
			}
			evidence = append(evidence, model.EvidenceFromBlock(strings.TrimRight(b.String(), "\n"), rootCause)...)
		}
		recommendation = buildRecommendation(ctx.ContainerConfigErrors[0], ctx.Pod.Namespace)
	} else {
		// Basic evidence (sweep mode)
		for _, cs := range ctx.Pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil && cs.State.Waiting.Reason == "CreateContainerConfigError" {
				evidence = append(evidence, model.EvidenceLine(fmt.Sprintf("Container: %s", cs.Name), false))
				evidence = append(evidence, model.EvidenceLine(fmt.Sprintf("Error: %s", cs.State.Waiting.Message), false))
			}
		}
		recommendation = "Verify secrets, configmaps, and volume mounts exist in the namespace. Check: kubectl get secrets,configmaps -n " + ctx.Pod.Namespace
	}

	if len(evidence) == 0 {
		return nil
	}
	return []model.Issue{{
		ID:             constants.IssueIDCreateContainerConfigError,
		Severity:       constants.SeverityCritical,
		ResourceKind:   "Pod",
		ResourceName:   resourceName,
		Message:        "CreateContainerConfigError detected",
		Evidence:       evidence,
		Recommendation: recommendation,
	}}
}

func buildRecommendation(err model.ContainerConfigErrorInfo, ns string) string {
	if err.Exists {
		return fmt.Sprintf("%s %s exists — verify key references and RBAC.\n\nkubectl describe %s %s -n %s",
			err.ResourceType, err.ResourceName, strings.ToLower(err.ResourceType), err.ResourceName, ns)
	}
	if err.ResourceType == "ConfigMap" {
		return "Create the missing ConfigMap or correct the reference."
	}
	if err.ResourceType == "Secret" {
		return fmt.Sprintf("Create the missing Secret or correct the reference in the pod spec.\n\nkubectl create secret generic %s -n %s ...", err.ResourceName, ns)
	}
	return fmt.Sprintf("Create the missing %s %s in namespace %s, or fix the reference in the pod spec.", err.ResourceType, err.ResourceName, ns)
}
