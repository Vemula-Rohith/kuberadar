package utils

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IsCrashLoop returns true if the pod is in CrashLoopBackOff.
// Checks both app containers and init containers.
func IsCrashLoop(pod *v1.Pod) bool {
	_, _, ok := CrashLoopContainer(pod)
	return ok
}

// CrashLoopContainer returns the name and restart count of the first container
// in CrashLoopBackOff, or ("", 0, false) if none. Checks app and init containers.
func CrashLoopContainer(pod *v1.Pod) (name string, restartCount int32, ok bool) {
	if pod == nil {
		return "", 0, false
	}
	check := func(statuses []v1.ContainerStatus) (string, int32, bool) {
		for _, cs := range statuses {
			if cs.State.Waiting != nil &&
				cs.State.Waiting.Reason == "CrashLoopBackOff" {
				return cs.Name, cs.RestartCount, true
			}
		}
		return "", 0, false
	}
	if name, count, ok := check(pod.Status.ContainerStatuses); ok {
		return name, count, true
	}
	return check(pod.Status.InitContainerStatuses)
}

// GetContainerExitCodes returns exit codes of terminated containers.
func GetContainerExitCodes(pod *v1.Pod) []int32 {
	var codes []int32
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Terminated != nil {
			codes = append(codes, cs.State.Terminated.ExitCode)
		}
		if cs.LastTerminationState.Terminated != nil {
			codes = append(codes, cs.LastTerminationState.Terminated.ExitCode)
		}
	}
	return codes
}

// TerminationInfo holds details from a container's last termination.
type TerminationInfo struct {
	ExitCode   int32
	Reason     string
	Message    string
	FinishedAt metav1.Time
}

// GetLastTerminationInfo returns termination info for the crashing container, if any.
func GetLastTerminationInfo(pod *v1.Pod, containerName string) (TerminationInfo, bool) {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Name != containerName {
			continue
		}
		if cs.LastTerminationState.Terminated != nil {
			t := cs.LastTerminationState.Terminated
			return TerminationInfo{
				ExitCode:   t.ExitCode,
				Reason:     t.Reason,
				Message:    t.Message,
				FinishedAt: t.FinishedAt,
			}, true
		}
	}
	// Check init containers
	for _, cs := range pod.Status.InitContainerStatuses {
		if cs.Name != containerName {
			continue
		}
		if cs.LastTerminationState.Terminated != nil {
			t := cs.LastTerminationState.Terminated
			return TerminationInfo{
				ExitCode:   t.ExitCode,
				Reason:     t.Reason,
				Message:    t.Message,
				FinishedAt: t.FinishedAt,
			}, true
		}
	}
	return TerminationInfo{}, false
}

// IsImagePullBackoff returns true if the pod has ImagePullBackOff.
func IsImagePullBackoff(pod *v1.Pod) bool {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil && cs.State.Waiting.Reason == "ImagePullBackOff" {
			return true
		}
		if cs.State.Waiting != nil && cs.State.Waiting.Reason == "ErrImagePull" {
			return true
		}
	}
	return false
}

// OrDefault returns s if non-empty, else d.
func OrDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

// IsCreateContainerConfigError returns true if the pod has CreateContainerConfigError.
func IsCreateContainerConfigError(pod *v1.Pod) bool {
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.State.Waiting != nil && cs.State.Waiting.Reason == "CreateContainerConfigError" {
			return true
		}
	}
	return false
}

// IsUnschedulable returns true if the pod is unschedulable (Pending with reason).
func IsUnschedulable(pod *v1.Pod) bool {
	if pod.Status.Phase != v1.PodPending {
		return false
	}
	for _, cond := range pod.Status.Conditions {
		if cond.Type == v1.PodScheduled && cond.Status == v1.ConditionFalse && cond.Reason == "Unschedulable" {
			return true
		}
	}
	return false
}
