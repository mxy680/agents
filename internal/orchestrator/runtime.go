package orchestrator

import (
	"context"
	"io"
)

// ContainerRuntime abstracts container orchestration (K8s or Docker).
type ContainerRuntime interface {
	// RunContainer starts a container and returns its name.
	RunContainer(ctx context.Context, spec ContainerSpec) (string, error)
	// StopContainer stops a running container.
	StopContainer(ctx context.Context, name string) error
	// RemoveContainer removes a stopped container.
	RemoveContainer(ctx context.Context, name string) error
	// ContainerLogs returns a streaming reader for container stdout+stderr.
	ContainerLogs(ctx context.Context, name string) (io.ReadCloser, error)
	// ContainerStatus returns the current status of a container.
	ContainerStatus(ctx context.Context, name string) (string, error)
	// ListContainers returns all agent containers.
	ListContainers(ctx context.Context) ([]ContainerInfo, error)
}

// ContainerSpec defines what to run.
type ContainerSpec struct {
	Name        string
	Image       string
	Env         map[string]string
	Command     []string
	Labels      map[string]string
	MemoryLimit string // e.g. "1g"
	CPULimit    string // e.g. "1"
}

// ContainerInfo is a simplified container status report.
type ContainerInfo struct {
	Name   string
	Status string // "running", "stopped", "failed", "completed"
	Labels map[string]string
}
