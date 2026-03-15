package providers

import (
	"bytes"
	"context"
	"io"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PodProvider fetches pod resources.
type PodProvider struct {
	client kubernetes.Interface
}

// NewPodProvider creates a new PodProvider.
func NewPodProvider(client kubernetes.Interface) *PodProvider {
	return &PodProvider{client: client}
}

// ListPods returns pods in the given namespace. Use "" for all namespaces.
func (p *PodProvider) ListPods(ctx context.Context, namespace string) ([]v1.Pod, error) {
	list, err := p.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetPod returns a single pod by namespace and name.
func (p *PodProvider) GetPod(ctx context.Context, namespace, name string) (*v1.Pod, error) {
	return p.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

// ListPodsWithSelector returns pods matching the label selector in the given namespace.
func (p *PodProvider) ListPodsWithSelector(ctx context.Context, namespace, selector string) ([]v1.Pod, error) {
	list, err := p.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetContainerLogs fetches the last N lines from a container.
// When previous is true, fetches from the previous instance (after crash).
// Returns empty string on error.
func (p *PodProvider) GetContainerLogs(ctx context.Context, namespace, podName, containerName string, tailLines int64, previous bool) string {
	opts := &v1.PodLogOptions{
		Container: containerName,
		Previous:  previous,
		TailLines: &tailLines,
	}
	req := p.client.CoreV1().Pods(namespace).GetLogs(podName, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return ""
	}
	defer stream.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, stream); err != nil {
		return ""
	}
	return buf.String()
}

// GetLogsForCrashLoop tries previous logs first; if empty (e.g. restartCount=1),
// falls back to current container logs.
func (p *PodProvider) GetLogsForCrashLoop(ctx context.Context, namespace, podName, containerName string, tailLines int64) (logs string, fromPrevious bool) {
	logs = strings.TrimSpace(p.GetContainerLogs(ctx, namespace, podName, containerName, tailLines, true))
	if logs != "" {
		return logs, true
	}
	return strings.TrimSpace(p.GetContainerLogs(ctx, namespace, podName, containerName, tailLines, false)), false
}
