package providers

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// NodeProvider fetches node resources.
type NodeProvider struct {
	client kubernetes.Interface
}

// NewNodeProvider creates a new NodeProvider.
func NewNodeProvider(client kubernetes.Interface) *NodeProvider {
	return &NodeProvider{client: client}
}

// GetNode returns a node by name.
func (n *NodeProvider) GetNode(ctx context.Context, name string) (*v1.Node, error) {
	return n.client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
}

// ListNodes returns all nodes.
func (n *NodeProvider) ListNodes(ctx context.Context) ([]v1.Node, error) {
	list, err := n.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
