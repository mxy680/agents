package orchestrator

import (
	"context"
	"fmt"
	"io"
)

// K8sRuntime adapts *K8sClient to the ContainerRuntime interface.
// It uses BuildPodSpec to construct a full K8s pod (with init containers
// for credential injection) and maps K8s pod operations to the interface.
type K8sRuntime struct {
	k8s *K8sClient
	cfg Config
}

// NewK8sRuntime creates a K8sRuntime that wraps the given K8sClient.
func NewK8sRuntime(k8s *K8sClient, cfg Config) *K8sRuntime {
	return &K8sRuntime{k8s: k8s, cfg: cfg}
}

// RunContainer builds a K8s pod spec from ContainerSpec and creates the pod.
// Env vars in spec.Env are passed into the pod via the init container mechanism
// defined in BuildPodSpec, so credentials are never written directly to the
// main container's environment in the pod manifest.
func (r *K8sRuntime) RunContainer(ctx context.Context, spec ContainerSpec) (string, error) {
	// Extract agent-specific env vars from the flat map.
	// BuildPodSpec handles credential injection via an init container; the
	// remaining well-known vars are stitched in directly.
	instanceID, _ := spec.Labels["instance-id"]
	userID, _ := spec.Labels["user-id"]

	// Separate the CLAUDE_CODE_OAUTH_TOKEN and AGENT_* vars from credentials.
	// All other env vars are treated as integration credentials.
	creds := make(map[string]string, len(spec.Env))
	for k, v := range spec.Env {
		switch k {
		case "CLAUDE_CODE_OAUTH_TOKEN", "AGENT_INSTANCE_ID", "AGENT_TEMPLATE":
			// These are injected by BuildPodSpec separately.
		default:
			creds[k] = v
		}
	}

	// Find the AgentTemplate name from env to satisfy BuildPodSpec.
	templateName := spec.Env["AGENT_TEMPLATE"]

	pod := BuildPodSpec(PodSpecParams{
		InstanceID: instanceID,
		UserID:     userID,
		Template: AgentTemplate{
			Name:        templateName,
			DockerImage: spec.Image,
		},
		Namespace:   r.cfg.KubeNamespace,
		Credentials: creds,
		Config:      r.cfg,
	})

	created, err := r.k8s.CreatePod(ctx, pod)
	if err != nil {
		return "", fmt.Errorf("k8s runtime: create pod: %w", err)
	}
	return created.Name, nil
}

// StopContainer deletes the K8s pod (K8s has no separate stop/remove).
func (r *K8sRuntime) StopContainer(ctx context.Context, name string) error {
	if err := r.k8s.DeletePod(ctx, name); err != nil {
		return fmt.Errorf("k8s runtime: stop container: %w", err)
	}
	return nil
}

// RemoveContainer deletes the K8s pod if it still exists.
func (r *K8sRuntime) RemoveContainer(ctx context.Context, name string) error {
	return r.StopContainer(ctx, name)
}

// ContainerLogs returns a streaming reader of the pod's main container logs.
func (r *K8sRuntime) ContainerLogs(ctx context.Context, name string) (io.ReadCloser, error) {
	stream, err := r.k8s.GetPodLogs(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("k8s runtime: get logs: %w", err)
	}
	return stream, nil
}

// ContainerStatus returns the normalised status of a K8s pod.
func (r *K8sRuntime) ContainerStatus(ctx context.Context, name string) (string, error) {
	pod, err := r.k8s.GetPod(ctx, name)
	if err != nil {
		return "", fmt.Errorf("k8s runtime: get pod: %w", err)
	}
	return PodStatus(pod), nil
}

// ListContainers returns all pods labelled app=agent as ContainerInfo.
func (r *K8sRuntime) ListContainers(ctx context.Context) ([]ContainerInfo, error) {
	pods, err := r.k8s.ListAgentPods(ctx)
	if err != nil {
		return nil, fmt.Errorf("k8s runtime: list pods: %w", err)
	}

	infos := make([]ContainerInfo, 0, len(pods))
	for _, pod := range pods {
		infos = append(infos, ContainerInfo{
			Name:   pod.Name,
			Status: PodStatus(&pod),
			Labels: pod.Labels,
		})
	}
	return infos, nil
}
