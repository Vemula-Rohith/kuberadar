package detectors

import "github.com/Vemula-Rohith/kuberadar/internal/model"

// PodDetectors are detectors that apply to pod scope.
var PodDetectors = []Detector{
	CrashLoopDetector{},
	OOMKillDetector{},
	ImagePullDetector{},
	ContainerConfigDetector{},
	SchedulingDetector{},
	StaleConfigDetector{},
}

// DetectorsForScope returns detectors that match the given scope type.
func DetectorsForScope(scopeType model.ScopeType) []Detector {
	var out []Detector
	for _, d := range PodDetectors {
		if d.Target() == scopeType {
			out = append(out, d)
		}
	}
	return out
}
