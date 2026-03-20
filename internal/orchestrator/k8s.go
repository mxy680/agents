package orchestrator

import (
	"context"
	"fmt"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sClient wraps Kubernetes API operations for agent pod management.
type K8sClient struct {
	clientset kubernetes.Interface
	namespace string
}

// NewK8sClient creates a K8sClient using in-cluster config or kubeconfig.
func NewK8sClient(namespace string) (*K8sClient, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig for local development
		config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		if err != nil {
			return nil, fmt.Errorf("k8s config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("k8s clientset: %w", err)
	}

	return &K8sClient{clientset: clientset, namespace: namespace}, nil
}

// NewK8sClientFromClientset creates a K8sClient from an existing clientset (for testing).
func NewK8sClientFromClientset(cs kubernetes.Interface, namespace string) *K8sClient {
	return &K8sClient{clientset: cs, namespace: namespace}
}

// CreatePod creates a pod in the configured namespace.
func (k *K8sClient) CreatePod(ctx context.Context, pod *corev1.Pod) (*corev1.Pod, error) {
	created, err := k.clientset.CoreV1().Pods(k.namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("create pod: %w", err)
	}
	return created, nil
}

// DeletePod deletes a pod by name.
func (k *K8sClient) DeletePod(ctx context.Context, name string) error {
	err := k.clientset.CoreV1().Pods(k.namespace).Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("delete pod %s: %w", name, err)
	}
	return nil
}

// GetPod returns a pod by name.
func (k *K8sClient) GetPod(ctx context.Context, name string) (*corev1.Pod, error) {
	pod, err := k.clientset.CoreV1().Pods(k.namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get pod %s: %w", name, err)
	}
	return pod, nil
}

// GetPodLogs returns the logs for a pod's main container.
func (k *K8sClient) GetPodLogs(ctx context.Context, name string) (io.ReadCloser, error) {
	req := k.clientset.CoreV1().Pods(k.namespace).GetLogs(name, &corev1.PodLogOptions{
		Container: "agent",
		Follow:    true,
	})
	stream, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("pod logs %s: %w", name, err)
	}
	return stream, nil
}

// ListAgentPods returns all pods with the app=agent label.
func (k *K8sClient) ListAgentPods(ctx context.Context) ([]corev1.Pod, error) {
	list, err := k.clientset.CoreV1().Pods(k.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=agent",
	})
	if err != nil {
		return nil, fmt.Errorf("list agent pods: %w", err)
	}
	return list.Items, nil
}

// PodStatus returns a simplified status string from a K8s Pod.
func PodStatus(pod *corev1.Pod) string {
	switch pod.Status.Phase {
	case corev1.PodPending:
		return StatusCreating
	case corev1.PodRunning:
		return StatusRunning
	case corev1.PodSucceeded:
		return StatusCompleted
	case corev1.PodFailed:
		return StatusFailed
	default:
		return StatusFailed
	}
}
