package model

import v1 "k8s.io/api/core/v1"

// ContainerLogs holds log content and whether it came from the previous instance.
type ContainerLogs struct {
	Content      string
	FromPrevious bool
}

// ContainerConfigErrorInfo holds verified info about CreateContainerConfigError.
type ContainerConfigErrorInfo struct {
	ResourceType string   // "ConfigMap" or "Secret"
	ResourceName string
	Exists       bool     // verified via API
	ReferencedBy []string // e.g. "Volume: x, MountPath: /etc/config"
}

// PVCErrorInfo holds verified info about unschedulable pod due to missing/unbound PVC.
type PVCErrorInfo struct {
	ClaimName           string   // PVC name
	Exists              bool     // verified via API
	IsBound             bool     // true if PVC exists and Phase == ClaimBound
	StorageClassName    string   // from pvc.Spec.StorageClassName (empty if not set)
	StorageClassExists  bool     // true if StorageClass exists in cluster (or not specified)
	ReferencedBy        []string // e.g. "Container: app\nVolume: data\nMountPath: /data"
}

// Context holds the Kubernetes resources needed for detection.
// Detectors receive this; they do not make API calls.
type Context struct {
	Pod                    *v1.Pod
	Events                 []v1.Event
	Node                   *v1.Node
	Namespace              string
	ContainerLogs          map[string]ContainerLogs       // container name -> logs (only populated in diagnose mode)
	ContainerConfigErrors  []ContainerConfigErrorInfo     // verified CreateContainerConfigError info (only in diagnose mode)
	SchedulingErrors       []PVCErrorInfo                 // verified PVC-not-found info for unschedulable pods (only in diagnose mode)
}
