package providers

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// StorageClassProvider checks if StorageClasses exist (cluster-scoped).
type StorageClassProvider struct {
	client kubernetes.Interface
}

// NewStorageClassProvider creates a new StorageClassProvider.
func NewStorageClassProvider(client kubernetes.Interface) *StorageClassProvider {
	return &StorageClassProvider{client: client}
}

// StorageClassExists returns true if the StorageClass exists in the cluster.
func (p *StorageClassProvider) StorageClassExists(ctx context.Context, name string) bool {
	if name == "" {
		return true // no StorageClass specified, cluster default applies
	}
	_, err := p.client.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
	return err == nil
}
