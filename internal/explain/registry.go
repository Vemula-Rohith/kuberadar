package explain

import "github.com/Vemula-Rohith/kuberadar/internal/constants"

// Entry holds documentation for an issue ID.
type Entry struct {
	ID                string
	Name              string
	Description       string
	CommonCauses      []string
	RecommendedActions []string
}

// Registry maps issue IDs to their documentation.
var Registry = map[string]Entry{
	constants.IssueIDOOMKilled: {
		ID:          constants.IssueIDOOMKilled,
		Name:        "OOMKilled",
		Description: "The container was terminated because it exceeded its memory limit.",
		CommonCauses: []string{
			"memory limit too low for workload",
			"memory leak in application",
			"spike in traffic or data volume",
		},
		RecommendedActions: []string{
			"increase memory limits or requests",
			"profile application memory usage",
			"consider horizontal pod autoscaling",
		},
	},
	constants.IssueIDCrashLoopBackOff: {
		ID:          constants.IssueIDCrashLoopBackOff,
		Name:        "CrashLoopBackOff",
		Description: "The container is repeatedly crashing and Kubernetes is backing off before restarting it.",
		CommonCauses: []string{
			"application error or unhandled exception",
			"missing configuration or environment variables",
			"failed health checks or startup probes",
			"insufficient resources",
		},
		RecommendedActions: []string{
			"check container logs: kubectl logs <pod>",
			"check previous container logs: kubectl logs <pod> --previous",
			"verify environment variables and config",
			"fix the application error or add proper error handling",
		},
	},
	constants.IssueIDImagePullBackOff: {
		ID:          constants.IssueIDImagePullBackOff,
		Name:        "ImagePullBackOff",
		Description: "Kubernetes cannot pull the container image.",
		CommonCauses: []string{
			"incorrect image tag",
			"private registry authentication failure",
			"registry outage",
			"network connectivity issues",
		},
		RecommendedActions: []string{
			"verify image exists: docker pull <image>",
			"check imagePullSecrets for private registries",
			"verify image name and tag are correct",
			"test registry accessibility from cluster",
		},
	},
	constants.IssueIDUnschedulablePod: {
		ID:          constants.IssueIDUnschedulablePod,
		Name:        "UnschedulablePod",
		Description: "The pod cannot be scheduled onto any node in the cluster.",
		CommonCauses: []string{
			"insufficient CPU or memory on nodes",
			"node taints not tolerated by pod",
			"pod affinity/anti-affinity rules",
			"PVC cannot be bound",
			"no nodes match node selector",
		},
		RecommendedActions: []string{
			"check node resources: kubectl describe nodes",
			"review taints and tolerations",
			"verify PVC exists and is bound",
			"check scheduler events: kubectl get events",
		},
	},
	constants.IssueIDCreateContainerConfigError: {
		ID:          constants.IssueIDCreateContainerConfigError,
		Name:        "CreateContainerConfigError",
		Description: "Kubernetes cannot create the container configuration. Typically caused by missing secrets, configmaps, or invalid volume mounts.",
		CommonCauses: []string{
			"secret not found",
			"configmap not found",
			"invalid volume mount path",
			"downwardAPI or projected volume misconfiguration",
		},
		RecommendedActions: []string{
			"verify referenced secrets exist: kubectl get secret <name> -n <namespace>",
			"verify referenced configmaps exist: kubectl get configmap <name> -n <namespace>",
			"check pod spec for typos in secret/configmap names",
			"describe the pod: kubectl describe pod <name> -n <namespace>",
		},
	},
}
