package model

// ScopeType defines the type of diagnostic scope.
type ScopeType string

const (
	ScopePod        ScopeType = "pod"
	ScopeDeployment ScopeType = "deployment"
	ScopeNamespace  ScopeType = "namespace"
	ScopeCluster    ScopeType = "cluster"
)

// Scope defines the scope for a diagnostic run.
type Scope struct {
	Type      ScopeType
	Namespace string
	Name      string
	Diagnose  bool // true for deep investigation (e.g. fetch logs); never set during sweep
}
