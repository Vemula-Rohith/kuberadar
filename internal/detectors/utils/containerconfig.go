package utils

import (
	"regexp"
	"strings"

	v1 "k8s.io/api/core/v1"
)

// ParsedConfigError holds extracted info from CreateContainerConfigError message.
type ParsedConfigError struct {
	ResourceType string // "ConfigMap" or "Secret"
	ResourceName string
	RawMessage   string
}

// ParseCreateContainerConfigError extracts resource type and name from the error message.
// Handles: configmap "x" not found, secret "x" not found, secret 'x' not found
func ParseCreateContainerConfigError(msg string) (ParsedConfigError, bool) {
	if msg == "" {
		return ParsedConfigError{}, false
	}
	msg = strings.TrimSpace(msg)
	// Match: configmap "name" or configmap 'name' or secret "name" or secret 'name'
	re := regexp.MustCompile(`(?i)(configmap|secret)\s+["']([^"']+)["']`)
	m := re.FindStringSubmatch(msg)
	if len(m) != 3 {
		return ParsedConfigError{RawMessage: msg}, false
	}
	rt := strings.ToLower(m[1])
	if rt == "configmap" {
		rt = "ConfigMap"
	} else if rt == "secret" {
		rt = "Secret"
	}
	return ParsedConfigError{
		ResourceType: rt,
		ResourceName: m[2],
		RawMessage:   msg,
	}, true
}

// ConfigRef describes where a secret/configmap is referenced in the pod.
type ConfigRef struct {
	ContainerName string
	VolumeName    string
	MountPath     string
	EnvVar        string
}

// FindConfigReferences returns where the given secret/configmap name is used in the pod.
func FindConfigReferences(pod *v1.Pod, resourceName string) []ConfigRef {
	if pod == nil {
		return nil
	}
	var refs []ConfigRef
	for _, vol := range pod.Spec.Volumes {
		if vol.ConfigMap != nil && vol.ConfigMap.Name == resourceName {
			for _, c := range pod.Spec.Containers {
				for _, m := range c.VolumeMounts {
					if m.Name == vol.Name {
						refs = append(refs, ConfigRef{
							ContainerName: c.Name,
							VolumeName:    vol.Name,
							MountPath:     m.MountPath,
						})
						break
					}
				}
			}
		}
		if vol.Secret != nil && vol.Secret.SecretName == resourceName {
			for _, c := range pod.Spec.Containers {
				for _, m := range c.VolumeMounts {
					if m.Name == vol.Name {
						refs = append(refs, ConfigRef{
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
	for _, c := range pod.Spec.Containers {
		for _, e := range c.Env {
			if e.ValueFrom != nil {
				if e.ValueFrom.ConfigMapKeyRef != nil && e.ValueFrom.ConfigMapKeyRef.Name == resourceName {
					refs = append(refs, ConfigRef{ContainerName: c.Name, EnvVar: e.Name})
				}
				if e.ValueFrom.SecretKeyRef != nil && e.ValueFrom.SecretKeyRef.Name == resourceName {
					refs = append(refs, ConfigRef{ContainerName: c.Name, EnvVar: e.Name})
				}
			}
		}
		for _, e := range c.EnvFrom {
			if e.ConfigMapRef != nil && e.ConfigMapRef.Name == resourceName {
				refs = append(refs, ConfigRef{ContainerName: c.Name, EnvVar: "envFrom: " + e.Prefix + "*"})
			}
			if e.SecretRef != nil && e.SecretRef.Name == resourceName {
				refs = append(refs, ConfigRef{ContainerName: c.Name, EnvVar: "envFrom: " + e.Prefix + "*"})
			}
		}
	}
	return refs
}
