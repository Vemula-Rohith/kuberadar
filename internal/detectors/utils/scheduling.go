package utils

import (
	"regexp"
	"strings"

	v1 "k8s.io/api/core/v1"
)

// ParseSchedulerMessageForPVC extracts PVC name from scheduler message.
// Handles: persistentvolumeclaim "x" not found, persistentvolumeclaim 'x' not found
func ParseSchedulerMessageForPVC(msg string) (claimName string, ok bool) {
	if msg == "" {
		return "", false
	}
	re := regexp.MustCompile(`(?i)persistentvolumeclaim\s+["']([^"']+)["']`)
	m := re.FindStringSubmatch(msg)
	if len(m) != 2 {
		return "", false
	}
	return strings.TrimSpace(m[1]), true
}

// SchedulerMessageIndicatesUnboundPVC returns true if the message indicates unbound PVCs.
// Handles: "pod has unbound immediate PersistentVolumeClaims"
func SchedulerMessageIndicatesUnboundPVC(msg string) bool {
	return strings.Contains(strings.ToLower(msg), "unbound") &&
		strings.Contains(strings.ToLower(msg), "persistentvolumeclaim")
}

// ListPodPVCClaimNames returns all PVC claim names referenced in the pod.
func ListPodPVCClaimNames(pod *v1.Pod) []string {
	if pod == nil {
		return nil
	}
	seen := make(map[string]bool)
	var names []string
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName != "" {
			if !seen[vol.PersistentVolumeClaim.ClaimName] {
				seen[vol.PersistentVolumeClaim.ClaimName] = true
				names = append(names, vol.PersistentVolumeClaim.ClaimName)
			}
		}
	}
	return names
}

// PVCRef describes where a PVC is referenced in the pod.
type PVCRef struct {
	ContainerName string
	VolumeName    string
	MountPath     string
}

// FindPVCReferences returns where the given PVC claim name is used in the pod.
func FindPVCReferences(pod *v1.Pod, claimName string) []PVCRef {
	if pod == nil {
		return nil
	}
	var refs []PVCRef
	for _, vol := range pod.Spec.Volumes {
		if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName == claimName {
			for _, c := range pod.Spec.Containers {
				for _, m := range c.VolumeMounts {
					if m.Name == vol.Name {
						refs = append(refs, PVCRef{
							ContainerName: c.Name,
							VolumeName:    vol.Name,
							MountPath:     m.MountPath,
						})
						break
					}
				}
			}
		}
	}
	return refs
}
