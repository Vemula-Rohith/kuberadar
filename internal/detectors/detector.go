package detectors

import "github.com/Vemula-Rohith/kuberadar/internal/model"

// Detector defines the interface for diagnostic rules.
type Detector interface {
	Name() string
	Target() model.ScopeType
	Detect(ctx model.Context) []model.Issue
}
