package engine

import (
	"context"
	"strconv"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Vemula-Rohith/kuberadar/internal/detectors"
	"github.com/Vemula-Rohith/kuberadar/internal/model"
	"github.com/Vemula-Rohith/kuberadar/internal/providers"
)

// Engine runs diagnostics for a given scope.
type Engine struct {
	podProvider    *providers.PodProvider
	deployProvider *providers.DeploymentProvider
	configProvider *providers.ConfigProvider
	contextBuilder *ContextBuilder
}

// NewEngine creates a new Engine.
func NewEngine(podProvider *providers.PodProvider, deployProvider *providers.DeploymentProvider, configProvider *providers.ConfigProvider, contextBuilder *ContextBuilder) *Engine {
	return &Engine{
		podProvider:    podProvider,
		deployProvider: deployProvider,
		configProvider: configProvider,
		contextBuilder: contextBuilder,
	}
}

// RunOpts optional hooks for a Run.
type RunOpts struct {
	// OnProgress is called with short status text (e.g. for a terminal spinner).
	OnProgress func(phase string)
}

// Run executes diagnostics for the given scope and returns a Diagnosis.
// opts may be nil.
func (e *Engine) Run(ctx context.Context, scope model.Scope, opts *RunOpts) (*model.Diagnosis, error) {
	report := func(msg string) {
		if opts != nil && opts.OnProgress != nil {
			opts.OnProgress(msg)
		}
	}

	diagnosis := &model.Diagnosis{
		Scope:     scope,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	detectorList := detectors.DetectorsForScope(model.ScopePod)
	if len(detectorList) == 0 {
		return diagnosis, nil
	}

	report("Scanning cluster...")
	pods, err := e.fetchPods(ctx, scope)
	if err != nil {
		return nil, err
	}

	var configMapMap map[string]v1.ConfigMap
	var secretMap map[string]v1.Secret
	if e.configProvider != nil && scope.Namespace != "" {
		report("Loading ConfigMaps and Secrets...")
		if m, err := e.configProvider.ListConfigMapsMap(ctx, scope.Namespace); err == nil {
			configMapMap = m
		}
		if m, err := e.configProvider.ListSecretsMap(ctx, scope.Namespace); err == nil {
			secretMap = m
		}
	}

	// One map per Run: avoids O(pods) Node GETs when many pods share few nodes.
	nodeCache := make(map[string]*v1.Node)

	for i := range pods {
		pod := pods[i]
		if scope.Diagnose && i == 0 {
			report("Gathering context...")
			report("Checking events...")
		}
		if len(pods) > 1 {
			report("Checking pods... (" + itoa(i+1) + "/" + itoa(len(pods)) + ")")
		} else if !scope.Diagnose {
			report("Checking pods...")
		}
		mctx, err := e.contextBuilder.BuildPodContext(ctx, pod, scope.Namespace, scope.Diagnose, nodeCache, configMapMap, secretMap)
		if err != nil {
			continue
		}
		for _, d := range detectorList {
			issues := d.Detect(mctx)
			for j := range issues {
				if issues[j].ResourceKind == "" {
					issues[j].ResourceKind = "Pod"
				}
				if issues[j].ResourceName == "" {
					issues[j].ResourceName = pod.Namespace + "/" + pod.Name
				}
			}
			diagnosis.Add(issues...)
		}
		diagnosis.PodsScanned++
	}

	return diagnosis, nil
}

func (e *Engine) fetchPods(ctx context.Context, scope model.Scope) ([]v1.Pod, error) {
	switch scope.Type {
	case model.ScopeDeployment:
		if scope.Name == "" {
			return e.podProvider.ListPods(ctx, scope.Namespace)
		}
		return e.fetchPodsForDeployment(ctx, scope.Namespace, scope.Name)
	case model.ScopePod:
		if scope.Name != "" {
			pod, err := e.podProvider.GetPod(ctx, scope.Namespace, scope.Name)
			if err != nil {
				return nil, err
			}
			return []v1.Pod{*pod}, nil
		}
	}
	return e.podProvider.ListPods(ctx, scope.Namespace)
}

func (e *Engine) fetchPodsForDeployment(ctx context.Context, namespace, deployName string) ([]v1.Pod, error) {
	deploy, err := e.deployProvider.GetDeployment(ctx, namespace, deployName)
	if err != nil {
		return nil, err
	}
	selector, err := metav1.LabelSelectorAsSelector(deploy.Spec.Selector)
	if err != nil {
		return nil, err
	}
	return e.podProvider.ListPodsWithSelector(ctx, namespace, selector.String())
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
