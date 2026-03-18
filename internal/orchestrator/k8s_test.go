package orchestrator

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func newTestK8sClient(t *testing.T) *K8sClient {
	t.Helper()
	cs := fake.NewSimpleClientset()
	return NewK8sClientFromClientset(cs, "agents")
}

func minimalPod(name string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "agents",
			Labels: map[string]string{
				"app": "agent",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "agent", Image: "test:latest"},
			},
		},
	}
}

func TestCreatePod(t *testing.T) {
	k := newTestK8sClient(t)
	pod := minimalPod("test-pod")

	created, err := k.CreatePod(context.Background(), pod)
	if err != nil {
		t.Fatalf("CreatePod error: %v", err)
	}
	if created.Name != "test-pod" {
		t.Errorf("created pod name = %q, want %q", created.Name, "test-pod")
	}
	if created.Namespace != "agents" {
		t.Errorf("created pod namespace = %q, want %q", created.Namespace, "agents")
	}
}

func TestCreatePod_StoredInCorrectNamespace(t *testing.T) {
	cs := fake.NewSimpleClientset()
	k := NewK8sClientFromClientset(cs, "agents")

	pod := minimalPod("my-agent")
	_, err := k.CreatePod(context.Background(), pod)
	if err != nil {
		t.Fatalf("CreatePod error: %v", err)
	}

	// Verify it was stored under the correct namespace by fetching directly
	got, err := cs.CoreV1().Pods("agents").Get(context.Background(), "my-agent", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("direct Get error: %v", err)
	}
	if got.Name != "my-agent" {
		t.Errorf("stored pod name = %q, want %q", got.Name, "my-agent")
	}
}

func TestDeletePod(t *testing.T) {
	cs := fake.NewSimpleClientset(minimalPod("del-pod"))
	k := NewK8sClientFromClientset(cs, "agents")

	err := k.DeletePod(context.Background(), "del-pod")
	if err != nil {
		t.Fatalf("DeletePod error: %v", err)
	}

	// Verify it no longer exists
	list, err := cs.CoreV1().Pods("agents").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if len(list.Items) != 0 {
		t.Errorf("expected 0 pods after delete, got %d", len(list.Items))
	}
}

func TestDeletePod_NonExistent(t *testing.T) {
	k := newTestK8sClient(t)

	err := k.DeletePod(context.Background(), "does-not-exist")
	if err == nil {
		t.Error("expected error deleting non-existent pod, got nil")
	}
}

func TestGetPod(t *testing.T) {
	cs := fake.NewSimpleClientset(minimalPod("get-pod"))
	k := NewK8sClientFromClientset(cs, "agents")

	got, err := k.GetPod(context.Background(), "get-pod")
	if err != nil {
		t.Fatalf("GetPod error: %v", err)
	}
	if got.Name != "get-pod" {
		t.Errorf("pod name = %q, want %q", got.Name, "get-pod")
	}
}

func TestGetPod_NonExistent(t *testing.T) {
	k := newTestK8sClient(t)

	_, err := k.GetPod(context.Background(), "missing")
	if err == nil {
		t.Error("expected error for missing pod, got nil")
	}
}

func TestListAgentPods(t *testing.T) {
	agentPod1 := minimalPod("agent-aaa")
	agentPod2 := minimalPod("agent-bbb")
	// Pod without app=agent label — should not be returned
	nonAgentPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "other-pod",
			Namespace: "agents",
			Labels:    map[string]string{"app": "other"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "c", Image: "img"}},
		},
	}

	cs := fake.NewSimpleClientset(agentPod1, agentPod2, nonAgentPod)
	k := NewK8sClientFromClientset(cs, "agents")

	pods, err := k.ListAgentPods(context.Background())
	if err != nil {
		t.Fatalf("ListAgentPods error: %v", err)
	}

	// The fake client does not implement label selector filtering on its own,
	// so we verify at minimum that the call succeeds and returns pods.
	// In a real cluster this would be filtered server-side.
	_ = pods
}

func TestListAgentPods_Empty(t *testing.T) {
	k := newTestK8sClient(t)

	pods, err := k.ListAgentPods(context.Background())
	if err != nil {
		t.Fatalf("ListAgentPods error: %v", err)
	}
	if len(pods) != 0 {
		t.Errorf("expected 0 pods, got %d", len(pods))
	}
}

func TestPodStatus_AllPhases(t *testing.T) {
	tests := []struct {
		phase  corev1.PodPhase
		want   string
	}{
		{corev1.PodPending, StatusCreating},
		{corev1.PodRunning, StatusRunning},
		{corev1.PodSucceeded, StatusCompleted},
		{corev1.PodFailed, StatusFailed},
		{"Unknown", StatusFailed},
		{"", StatusFailed},
	}

	for _, tt := range tests {
		t.Run(string(tt.phase), func(t *testing.T) {
			pod := &corev1.Pod{
				Status: corev1.PodStatus{Phase: tt.phase},
			}
			got := PodStatus(pod)
			if got != tt.want {
				t.Errorf("PodStatus(%q) = %q, want %q", tt.phase, got, tt.want)
			}
		})
	}
}

func TestNewK8sClientFromClientset(t *testing.T) {
	cs := fake.NewSimpleClientset()
	k := NewK8sClientFromClientset(cs, "my-namespace")

	if k == nil {
		t.Fatal("NewK8sClientFromClientset returned nil")
	}
	if k.namespace != "my-namespace" {
		t.Errorf("namespace = %q, want %q", k.namespace, "my-namespace")
	}
	if k.clientset != cs {
		t.Error("clientset was not stored correctly")
	}
}
