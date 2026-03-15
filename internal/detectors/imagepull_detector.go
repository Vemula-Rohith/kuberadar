package detectors

import (
	"fmt"
	"strings"

	"github.com/kuberadar/kuberadar/internal/constants"
	"github.com/kuberadar/kuberadar/internal/detectors/utils"
	"github.com/kuberadar/kuberadar/internal/model"
)

// ImagePullDetector detects ImagePullBackOff and ErrImagePull.
type ImagePullDetector struct{}

// Name returns the detector name.
func (ImagePullDetector) Name() string {
	return "imagepull"
}

// Target returns the scope type this detector applies to.
func (ImagePullDetector) Target() model.ScopeType {
	return model.ScopePod
}

// Detect runs the detection logic.
func (d ImagePullDetector) Detect(ctx model.Context) []model.Issue {
	if ctx.Pod == nil {
		return nil
	}
	if !utils.IsImagePullBackoff(ctx.Pod) {
		return nil
	}
	imageByName := make(map[string]string)
	for _, c := range ctx.Pod.Spec.Containers {
		imageByName[c.Name] = c.Image
	}
	var evidence []model.Evidence
	var firstPullError bool
	for _, cs := range ctx.Pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil && (cs.State.Waiting.Reason == "ImagePullBackOff" || cs.State.Waiting.Reason == "ErrImagePull") {
			image := imageByName[cs.Name]
			evidence = append(evidence, model.EvidenceLine(fmt.Sprintf("Container: %s", cs.Name), false))
			evidence = append(evidence, model.EvidenceLine(fmt.Sprintf("Image: %s", image), false))
			evidence = append(evidence, model.EvidenceLine(fmt.Sprintf("Registry: %s", parseRegistry(image)), false))
			pullErrorLine := fmt.Sprintf("Pull error: %s", cs.State.Waiting.Message)
			evidence = append(evidence, model.EvidenceLine(pullErrorLine, !firstPullError))
			firstPullError = true
		}
	}
	resourceName := ctx.Pod.Namespace + "/" + ctx.Pod.Name
	return []model.Issue{{
		ID:             constants.IssueIDImagePullBackOff,
		Severity:       constants.SeverityCritical,
		ResourceKind:   "Pod",
		ResourceName:   resourceName,
		Message:        "ImagePullBackOff detected",
		Evidence:       evidence,
		Recommendation: "Verify image name, tag, and registry credentials. Check if the image exists and is accessible.",
	}}
}

func parseRegistry(image string) string {
	if image == "" {
		return "—"
	}
	// Format: [registry/]repo[:tag] — registry has a dot (gcr.io) or colon (host:5000)
	parts := strings.SplitN(image, "/", 2)
	if len(parts) < 2 {
		return "docker.io (default)"
	}
	first := parts[0]
	if strings.Contains(first, ".") || strings.Contains(first, ":") {
		return first // e.g. gcr.io, ghcr.io, registry.example.com:5000
	}
	return "docker.io"
}
