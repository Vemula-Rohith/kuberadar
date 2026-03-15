package providers

import (
	"context"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// EventProvider fetches event resources.
type EventProvider struct {
	client kubernetes.Interface
}

// NewEventProvider creates a new EventProvider.
func NewEventProvider(client kubernetes.Interface) *EventProvider {
	return &EventProvider{client: client}
}

// GetEventsForPod returns events for a pod in the given namespace.
func (e *EventProvider) GetEventsForPod(ctx context.Context, namespace, podName string) ([]v1.Event, error) {
	list, err := e.client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: "involvedObject.name=" + podName + ",involvedObject.kind=Pod",
	})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
