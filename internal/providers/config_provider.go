package providers

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ConfigProvider checks if ConfigMaps and Secrets exist.
type ConfigProvider struct {
	client kubernetes.Interface
}

// NewConfigProvider creates a new ConfigProvider.
func NewConfigProvider(client kubernetes.Interface) *ConfigProvider {
	return &ConfigProvider{client: client}
}

// ConfigMapExists returns true if the ConfigMap exists in the namespace.
func (p *ConfigProvider) ConfigMapExists(ctx context.Context, namespace, name string) bool {
	_, err := p.client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	return err == nil
}

// SecretExists returns true if the Secret exists in the namespace.
func (p *ConfigProvider) SecretExists(ctx context.Context, namespace, name string) bool {
	_, err := p.client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	return err == nil
}
