package engine

import (
	"context"
	"strings"

	v1 "k8s.io/api/core/v1"

	"github.com/Vemula-Rohith/kuberadar/internal/detectors/utils"
	"github.com/Vemula-Rohith/kuberadar/internal/model"
	"github.com/Vemula-Rohith/kuberadar/internal/providers"
)

// ContextBuilder builds model.Context for resources.
type ContextBuilder struct {
	eventProvider        *providers.EventProvider
	nodeProvider         *providers.NodeProvider
	podProvider          *providers.PodProvider
	configProvider       *providers.ConfigProvider
	pvcProvider          *providers.PVCProvider
	storageClassProvider *providers.StorageClassProvider
}

// NewContextBuilder creates a new ContextBuilder.
func NewContextBuilder(eventProvider *providers.EventProvider, nodeProvider *providers.NodeProvider, podProvider *providers.PodProvider, configProvider *providers.ConfigProvider, pvcProvider *providers.PVCProvider, storageClassProvider *providers.StorageClassProvider) *ContextBuilder {
	return &ContextBuilder{
		eventProvider:        eventProvider,
		nodeProvider:         nodeProvider,
		podProvider:          podProvider,
		configProvider:       configProvider,
		pvcProvider:          pvcProvider,
		storageClassProvider: storageClassProvider,
	}
}

// BuildPodContext fetches optional events, node, and diagnose-only data for a pod.
// nodeCache is keyed by node name; callers should pass a shared map across pods in one Run
// so each node is fetched at most once (major win for namespace sweeps).
// When diagnose is true and the pod is in CrashLoopBackOff, fetches previous container logs.
func (b *ContextBuilder) BuildPodContext(ctx context.Context, pod v1.Pod, namespace string, diagnose bool, nodeCache map[string]*v1.Node, configMaps map[string]v1.ConfigMap, secrets map[string]v1.Secret) (model.Context, error) {
	var events []v1.Event
	// Events are not consumed by any detector today; listing per pod is N API calls per sweep.
	// Fetch only in diagnose mode for future use / consistency with deep inspection.
	if diagnose && b.eventProvider != nil {
		var err error
		events, err = b.eventProvider.GetEventsForPod(ctx, pod.Namespace, pod.Name)
		if err != nil {
			events = nil
		}
	}
	var node *v1.Node
	if pod.Spec.NodeName != "" && b.nodeProvider != nil {
		node = lookupOrFetchNode(ctx, b.nodeProvider, pod.Spec.NodeName, nodeCache)
	}
	mctx := model.Context{
		Pod:        &pod,
		Events:     events,
		Node:       node,
		Namespace:  namespace,
		ConfigMaps: configMaps,
		Secrets:    secrets,
	}
	if diagnose && b.podProvider != nil {
		if containerName, _, ok := utils.CrashLoopContainer(&pod); ok {
			logs, fromPrevious := b.podProvider.GetLogsForCrashLoop(ctx, pod.Namespace, pod.Name, containerName, 20)
			mctx.ContainerLogs = map[string]model.ContainerLogs{
				containerName: {Content: logs, FromPrevious: fromPrevious},
			}
		}
	}
	if diagnose && b.configProvider != nil && utils.IsCreateContainerConfigError(&pod) {
		for _, cs := range pod.Status.ContainerStatuses {
			if cs.State.Waiting != nil && cs.State.Waiting.Reason == "CreateContainerConfigError" {
				parsed, ok := utils.ParseCreateContainerConfigError(cs.State.Waiting.Message)
				if ok {
					exists := false
					if parsed.ResourceType == "ConfigMap" {
						exists = b.configProvider.ConfigMapExists(ctx, pod.Namespace, parsed.ResourceName)
					} else if parsed.ResourceType == "Secret" {
						exists = b.configProvider.SecretExists(ctx, pod.Namespace, parsed.ResourceName)
					}
					refs := utils.FindConfigReferences(&pod, parsed.ResourceName)
					var refStrs []string
					for _, r := range refs {
						if r.VolumeName != "" {
							var lines []string
							if r.ContainerName != "" {
								lines = append(lines, "Container: "+r.ContainerName)
							}
							lines = append(lines, "Volume: "+r.VolumeName)
							if r.MountPath != "" {
								lines = append(lines, "MountPath: "+r.MountPath)
							}
							refStrs = append(refStrs, strings.Join(lines, "\n"))
						}
						if r.EnvVar != "" {
							var lines []string
							if r.ContainerName != "" {
								lines = append(lines, "Container: "+r.ContainerName)
							}
							if strings.HasPrefix(r.EnvVar, "envFrom:") {
								lines = append(lines, "Environment (envFrom): "+strings.TrimPrefix(r.EnvVar, "envFrom: "))
							} else {
								lines = append(lines, "Environment variable: "+r.EnvVar)
							}
							refStrs = append(refStrs, strings.Join(lines, "\n"))
						}
					}
					mctx.ContainerConfigErrors = append(mctx.ContainerConfigErrors, model.ContainerConfigErrorInfo{
						ResourceType: parsed.ResourceType,
						ResourceName: parsed.ResourceName,
						Exists:       exists,
						ReferencedBy: refStrs,
					})
				}
			}
		}
	}
	if diagnose && b.pvcProvider != nil && utils.IsUnschedulable(&pod) {
		for _, cond := range pod.Status.Conditions {
			if cond.Type == v1.PodScheduled && cond.Status == v1.ConditionFalse && cond.Message != "" {
				msg := cond.Message
				// Case 1: "persistentvolumeclaim \"x\" not found" — extract name from message
				if claimName, ok := utils.ParseSchedulerMessageForPVC(msg); ok {
					if info := buildPVCErrorInfo(ctx, b.pvcProvider, b.storageClassProvider, &pod, claimName); info != nil {
						mctx.SchedulingErrors = append(mctx.SchedulingErrors, *info)
					}
				} else if utils.SchedulerMessageIndicatesUnboundPVC(msg) {
					// Case 2: "pod has unbound immediate PersistentVolumeClaims" — check all pod PVCs
					for _, claimName := range utils.ListPodPVCClaimNames(&pod) {
						info := buildPVCErrorInfo(ctx, b.pvcProvider, b.storageClassProvider, &pod, claimName)
						if info != nil && info.Exists && !info.IsBound {
							mctx.SchedulingErrors = append(mctx.SchedulingErrors, *info)
						}
					}
				}
				break
			}
		}
	}
	return mctx, nil
}

func buildPVCErrorInfo(ctx context.Context, pvcProvider *providers.PVCProvider, scProvider *providers.StorageClassProvider, pod *v1.Pod, claimName string) *model.PVCErrorInfo {
	exists := pvcProvider.PVCExists(ctx, pod.Namespace, claimName)
	bound := false
	storageClassName := ""
	storageClassExists := true // default when not specified
	if exists {
		bound = pvcProvider.PVCIsBound(ctx, pod.Namespace, claimName)
		storageClassName = pvcProvider.GetPVCStorageClassName(ctx, pod.Namespace, claimName)
		if storageClassName != "" && scProvider != nil {
			storageClassExists = scProvider.StorageClassExists(ctx, storageClassName)
		}
	}
	refs := utils.FindPVCReferences(pod, claimName)
	var refStrs []string
	for _, r := range refs {
		var lines []string
		if r.ContainerName != "" {
			lines = append(lines, "Container: "+r.ContainerName)
		}
		lines = append(lines, "Volume: "+r.VolumeName)
		if r.MountPath != "" {
			lines = append(lines, "MountPath: "+r.MountPath)
		}
		refStrs = append(refStrs, strings.Join(lines, "\n"))
	}
	return &model.PVCErrorInfo{
		ClaimName:          claimName,
		Exists:             exists,
		IsBound:            bound,
		StorageClassName:   storageClassName,
		StorageClassExists: storageClassExists,
		ReferencedBy:       refStrs,
	}
}

// lookupOrFetchNode returns a cached node or fetches it once and stores in cache.
// Failed lookups are cached as nil to avoid repeated calls for missing nodes.
func lookupOrFetchNode(ctx context.Context, np *providers.NodeProvider, nodeName string, cache map[string]*v1.Node) *v1.Node {
	if cache != nil {
		if n, ok := cache[nodeName]; ok {
			return n
		}
	}
	n, err := np.GetNode(ctx, nodeName)
	if cache != nil {
		if err != nil {
			cache[nodeName] = nil
			return nil
		}
		cache[nodeName] = n
		return n
	}
	if err != nil {
		return nil
	}
	return n
}
