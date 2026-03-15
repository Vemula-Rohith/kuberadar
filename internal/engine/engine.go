package engine

import (
	"context"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kuberadar/kuberadar/internal/detectors"
	"github.com/kuberadar/kuberadar/internal/model"
	"github.com/kuberadar/kuberadar/internal/providers"
)

// Engine runs diagnostics for a given scope.
type Engine struct {
	podProvider       *providers.PodProvider
	deployProvider    *providers.DeploymentProvider
	contextBuilder    *ContextBuilder
}

// NewEngine creates a new Engine.
func NewEngine(podProvider *providers.PodProvider, deployProvider *providers.DeploymentProvider, contextBuilder *ContextBuilder) *Engine {
	return &Engine{
		podProvider:    podProvider,
		deployProvider: deployProvider,
		contextBuilder: contextBuilder,
	}
}

// Run executes diagnostics for the given scope and returns a Diagnosis.
func (e *Engine) Run(ctx context.Context, scope model.Scope) (*model.Diagnosis, error) {
	diagnosis := &model.Diagnosis{
		Scope:     scope,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	detectorList := detectors.DetectorsForScope(model.ScopePod)
	if len(detectorList) == 0 {
		return diagnosis, nil
	}

	pods, err := e.fetchPods(ctx, scope)
	if err != nil {
		return nil, err
	}

	for _, pod := range pods {
		mctx, err := e.contextBuilder.BuildPodContext(ctx, pod, scope.Namespace, scope.Diagnose)
		if err != nil {
			continue
		}
		for _, d := range detectorList {
			issues := d.Detect(mctx)
			for i := range issues {
				if issues[i].ResourceKind == "" {
					issues[i].ResourceKind = "Pod"
				}
				if issues[i].ResourceName == "" {
					issues[i].ResourceName = pod.Namespace + "/" + pod.Name
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
