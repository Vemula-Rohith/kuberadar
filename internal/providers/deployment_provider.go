package providers

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DeploymentProvider fetches deployment resources.
type DeploymentProvider struct {
	client kubernetes.Interface
}

// NewDeploymentProvider creates a new DeploymentProvider.
func NewDeploymentProvider(client kubernetes.Interface) *DeploymentProvider {
	return &DeploymentProvider{client: client}
}

// ListDeployments returns deployments in the given namespace.
func (d *DeploymentProvider) ListDeployments(ctx context.Context, namespace string) ([]appsv1.Deployment, error) {
	list, err := d.client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetDeployment returns a single deployment by namespace and name.
func (d *DeploymentProvider) GetDeployment(ctx context.Context, namespace, name string) (*appsv1.Deployment, error) {
	return d.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}
