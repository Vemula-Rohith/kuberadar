package utils

import (
	"sort"

	v1 "k8s.io/api/core/v1"
)

// GetConfigMapRefs returns deduplicated ConfigMap names referenced by the pod
// (envFrom, env valueFrom, volumes, projected volumes). Includes init and ephemeral containers.
func GetConfigMapRefs(pod *v1.Pod) []string {
	if pod == nil {
		return nil
	}
	set := make(map[string]struct{})
	for _, c := range append(append([]v1.Container{}, pod.Spec.InitContainers...), pod.Spec.Containers...) {
		addConfigMapRefsFromEnvFrom(c.EnvFrom, set)
		addConfigMapRefsFromEnv(c.Env, set)
	}
	for _, ec := range pod.Spec.EphemeralContainers {
		addConfigMapRefsFromEnvFrom(ec.EnvFrom, set)
		addConfigMapRefsFromEnv(ec.Env, set)
	}
	for _, vol := range pod.Spec.Volumes {
		if vol.ConfigMap != nil && vol.ConfigMap.Name != "" {
			set[vol.ConfigMap.Name] = struct{}{}
		}
		if vol.Projected == nil {
			continue
		}
		for _, src := range vol.Projected.Sources {
			if src.ConfigMap != nil && src.ConfigMap.Name != "" {
				set[src.ConfigMap.Name] = struct{}{}
			}
		}
	}
	return sortedStringSet(set)
}

// GetSecretRefs returns deduplicated Secret names referenced by the pod.
func GetSecretRefs(pod *v1.Pod) []string {
	if pod == nil {
		return nil
	}
	set := make(map[string]struct{})
	for _, c := range append(append([]v1.Container{}, pod.Spec.InitContainers...), pod.Spec.Containers...) {
		addSecretRefsFromEnvFrom(c.EnvFrom, set)
		addSecretRefsFromEnv(c.Env, set)
	}
	for _, ec := range pod.Spec.EphemeralContainers {
		addSecretRefsFromEnvFrom(ec.EnvFrom, set)
		addSecretRefsFromEnv(ec.Env, set)
	}
	for _, vol := range pod.Spec.Volumes {
		if vol.Secret != nil && vol.Secret.SecretName != "" {
			set[vol.Secret.SecretName] = struct{}{}
		}
		if vol.Projected == nil {
			continue
		}
		for _, src := range vol.Projected.Sources {
			if src.Secret != nil && src.Secret.Name != "" {
				set[src.Secret.Name] = struct{}{}
			}
		}
	}
	return sortedStringSet(set)
}

func addConfigMapRefsFromEnvFrom(src []v1.EnvFromSource, set map[string]struct{}) {
	for i := range src {
		if src[i].ConfigMapRef != nil && src[i].ConfigMapRef.Name != "" {
			set[src[i].ConfigMapRef.Name] = struct{}{}
		}
	}
}

func addConfigMapRefsFromEnv(env []v1.EnvVar, set map[string]struct{}) {
	for i := range env {
		if env[i].ValueFrom == nil || env[i].ValueFrom.ConfigMapKeyRef == nil {
			continue
		}
		if n := env[i].ValueFrom.ConfigMapKeyRef.Name; n != "" {
			set[n] = struct{}{}
		}
	}
}

func addSecretRefsFromEnvFrom(src []v1.EnvFromSource, set map[string]struct{}) {
	for i := range src {
		if src[i].SecretRef != nil && src[i].SecretRef.Name != "" {
			set[src[i].SecretRef.Name] = struct{}{}
		}
	}
}

func addSecretRefsFromEnv(env []v1.EnvVar, set map[string]struct{}) {
	for i := range env {
		if env[i].ValueFrom == nil || env[i].ValueFrom.SecretKeyRef == nil {
			continue
		}
		if n := env[i].ValueFrom.SecretKeyRef.Name; n != "" {
			set[n] = struct{}{}
		}
	}
}

func sortedStringSet(set map[string]struct{}) []string {
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
