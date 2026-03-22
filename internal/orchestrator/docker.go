package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// DockerRuntime implements ContainerRuntime by shelling out to the docker CLI.
type DockerRuntime struct{}

// NewDockerRuntime creates a new DockerRuntime.
func NewDockerRuntime() *DockerRuntime {
	return &DockerRuntime{}
}

// RunContainer starts a detached container and returns its name.
func (d *DockerRuntime) RunContainer(ctx context.Context, spec ContainerSpec) (string, error) {
	args := []string{"run", "-d", "--name=" + spec.Name, "--restart=no"}

	if spec.MemoryLimit != "" {
		args = append(args, "--memory="+spec.MemoryLimit)
	}
	if spec.CPULimit != "" {
		args = append(args, "--cpus="+spec.CPULimit)
	}

	for k, v := range spec.Labels {
		args = append(args, "--label="+k+"="+v)
	}

	for k, v := range spec.Env {
		args = append(args, "-e", k+"="+v)
	}

	args = append(args, spec.Image)
	args = append(args, spec.Command...)

	out, err := exec.CommandContext(ctx, "docker", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("docker run: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return spec.Name, nil
}

// StopContainer stops a running container.
func (d *DockerRuntime) StopContainer(ctx context.Context, name string) error {
	out, err := exec.CommandContext(ctx, "docker", "stop", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker stop %s: %w: %s", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// RemoveContainer removes a stopped container.
func (d *DockerRuntime) RemoveContainer(ctx context.Context, name string) error {
	out, err := exec.CommandContext(ctx, "docker", "rm", name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker rm %s: %w: %s", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// ContainerLogs returns a streaming reader for the container's stdout+stderr.
// The caller is responsible for closing the returned reader.
func (d *DockerRuntime) ContainerLogs(ctx context.Context, name string) (io.ReadCloser, error) {
	cmd := exec.CommandContext(ctx, "docker", "logs", "-f", name)
	// Combine stdout and stderr so agent output is fully captured.
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		pr.Close()
		pw.Close()
		return nil, fmt.Errorf("docker logs %s: %w", name, err)
	}

	// Close the write end of the pipe when the command exits so readers see EOF.
	go func() {
		cmd.Wait() //nolint:errcheck
		pw.Close()
	}()

	return pr, nil
}

// dockerInspectState is a minimal subset of docker inspect output.
type dockerInspectState struct {
	State struct {
		Status   string `json:"Status"`
		ExitCode int    `json:"ExitCode"`
	} `json:"State"`
}

// ContainerStatus returns the normalised status of a container.
func (d *DockerRuntime) ContainerStatus(ctx context.Context, name string) (string, error) {
	out, err := exec.CommandContext(ctx, "docker", "inspect", "--format={{json .}}", name).Output()
	if err != nil {
		return "", fmt.Errorf("docker inspect %s: %w", name, err)
	}

	var info dockerInspectState
	if err := json.Unmarshal(out, &info); err != nil {
		return "", fmt.Errorf("docker inspect parse %s: %w", name, err)
	}

	return normalizeDockerStatus(info.State.Status, info.State.ExitCode), nil
}

// normalizeDockerStatus maps a Docker state string to our status constants.
func normalizeDockerStatus(dockerStatus string, exitCode int) string {
	switch dockerStatus {
	case "running":
		return StatusRunning
	case "created":
		return StatusCreating
	case "exited":
		if exitCode == 0 {
			return StatusCompleted
		}
		return StatusFailed
	case "dead", "removing":
		return StatusFailed
	default:
		return StatusFailed
	}
}

// dockerPSEntry is a single record from docker ps --format json.
type dockerPSEntry struct {
	Names  string `json:"Names"`
	State  string `json:"State"`
	Labels string `json:"Labels"`
}

// ListContainers returns all containers labelled app=agent.
func (d *DockerRuntime) ListContainers(ctx context.Context) ([]ContainerInfo, error) {
	out, err := exec.CommandContext(ctx,
		"docker", "ps", "-a",
		"--filter", "label=app=agent",
		"--format", "{{json .}}",
	).Output()
	if err != nil {
		return nil, fmt.Errorf("docker ps: %w", err)
	}

	var results []ContainerInfo
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line == "" {
			continue
		}
		var entry dockerPSEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		// Parse label string "key=value,key2=value2" into a map.
		labels := parseDockerLabels(entry.Labels)

		results = append(results, ContainerInfo{
			Name:   strings.TrimPrefix(entry.Names, "/"),
			Status: normalizeDockerStatus(entry.State, 0),
			Labels: labels,
		})
	}
	return results, nil
}

// parseDockerLabels parses the comma-separated "key=value" label string
// that docker ps --format emits.
func parseDockerLabels(raw string) map[string]string {
	labels := make(map[string]string)
	for _, pair := range strings.Split(raw, ",") {
		k, v, found := strings.Cut(pair, "=")
		if found {
			labels[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return labels
}
