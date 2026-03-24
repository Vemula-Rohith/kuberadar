package utils

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ApproximateConfigResourceUpdateTime returns a best-effort last-change time for a
// ConfigMap or Secret. Prefers the newest ManagedFields entry time (server-side apply /
// controller updates); falls back to CreationTimestamp when empty.
func ApproximateConfigResourceUpdateTime(meta metav1.ObjectMeta) time.Time {
	var latest time.Time
	for i := range meta.ManagedFields {
		t := meta.ManagedFields[i].Time
		if t != nil && t.After(latest) {
			latest = t.Time
		}
	}
	if !latest.IsZero() {
		return latest
	}
	return meta.CreationTimestamp.Time
}
