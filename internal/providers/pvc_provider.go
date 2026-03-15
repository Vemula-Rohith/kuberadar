package providers

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PVCProvider checks if PersistentVolumeClaims exist and their binding status.
type PVCProvider struct {
	client kubernetes.Interface
}

// NewPVCProvider creates a new PVCProvider.
func NewPVCProvider(client kubernetes.Interface) *PVCProvider {
	return &PVCProvider{client: client}
}

// PVCExists returns true if the PVC exists in the namespace.
func (p *PVCProvider) PVCExists(ctx context.Context, namespace, name string) bool {
	_, err := p.client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
	return err == nil
}

// PVCIsBound returns true if the PVC exists and is bound to a PersistentVolume.
func (p *PVCProvider) PVCIsBound(ctx context.Context, namespace, name string) bool {
	pvc, err := p.client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return false
	}
	return pvc.Status.Phase == v1.ClaimBound
}

// GetPVCStorageClassName returns the StorageClass name from the PVC spec, or empty if not set.
func (p *PVCProvider) GetPVCStorageClassName(ctx context.Context, namespace, name string) string {
	pvc, err := p.client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return ""
	}
	if pvc.Spec.StorageClassName == nil || *pvc.Spec.StorageClassName == "" {
		return ""
	}
	return *pvc.Spec.StorageClassName
}
