package orchestrator

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func defaultTestParams() PodSpecParams {
	return PodSpecParams{
		InstanceID: "abcdef1234567890",
		UserID:     "user-42",
		Template: AgentTemplate{
			Name:        "email-responder",
			DockerImage: "ghcr.io/emdash/email-responder:v1",
		},
		Namespace: "agents",
		Credentials: map[string]string{
			"GOOGLE_ACCESS_TOKEN":  "tok-abc",
			"GOOGLE_REFRESH_TOKEN": "ref-xyz",
		},
		Config: Config{
			ExportCredsImage:       "ghcr.io/emdash/export-creds:latest",
			ClaudeSessionSecretRef: "claude-session-token",
		},
	}
}

func TestBuildPodSpec_PodName(t *testing.T) {
	p := defaultTestParams()
	pod := BuildPodSpec(p)

	expected := "agent-" + p.InstanceID[:8]
	if pod.Name != expected {
		t.Errorf("pod name = %q, want %q", pod.Name, expected)
	}
}

func TestBuildPodSpec_Namespace(t *testing.T) {
	p := defaultTestParams()
	pod := BuildPodSpec(p)

	if pod.Namespace != "agents" {
		t.Errorf("namespace = %q, want %q", pod.Namespace, "agents")
	}
}

func TestBuildPodSpec_Labels(t *testing.T) {
	p := defaultTestParams()
	pod := BuildPodSpec(p)

	labels := pod.Labels
	if labels["app"] != "agent" {
		t.Errorf("label app = %q, want %q", labels["app"], "agent")
	}
	if labels["instance-id"] != p.InstanceID {
		t.Errorf("label instance-id = %q, want %q", labels["instance-id"], p.InstanceID)
	}
	if labels["user-id"] != p.UserID {
		t.Errorf("label user-id = %q, want %q", labels["user-id"], p.UserID)
	}
}

func TestBuildPodSpec_RestartPolicy(t *testing.T) {
	pod := BuildPodSpec(defaultTestParams())

	if pod.Spec.RestartPolicy != corev1.RestartPolicyNever {
		t.Errorf("restart policy = %q, want %q", pod.Spec.RestartPolicy, corev1.RestartPolicyNever)
	}
}

func TestBuildPodSpec_InitContainerCredEnvVars(t *testing.T) {
	p := defaultTestParams()
	pod := BuildPodSpec(p)

	if len(pod.Spec.InitContainers) != 1 {
		t.Fatalf("init containers count = %d, want 1", len(pod.Spec.InitContainers))
	}
	init := pod.Spec.InitContainers[0]

	envMap := make(map[string]string)
	for _, ev := range init.Env {
		envMap[ev.Name] = ev.Value
	}

	for k, v := range p.Credentials {
		if got, ok := envMap[k]; !ok {
			t.Errorf("init container missing env var %q", k)
		} else if got != v {
			t.Errorf("init container env %q = %q, want %q", k, got, v)
		}
	}
}

func TestBuildPodSpec_InitContainerImage(t *testing.T) {
	p := defaultTestParams()
	pod := BuildPodSpec(p)

	init := pod.Spec.InitContainers[0]
	if init.Image != p.Config.ExportCredsImage {
		t.Errorf("init container image = %q, want %q", init.Image, p.Config.ExportCredsImage)
	}
}

func TestBuildPodSpec_InitContainerVolumeMount(t *testing.T) {
	pod := BuildPodSpec(defaultTestParams())
	init := pod.Spec.InitContainers[0]

	if len(init.VolumeMounts) != 1 {
		t.Fatalf("init volume mounts = %d, want 1", len(init.VolumeMounts))
	}
	if init.VolumeMounts[0].Name != "creds" {
		t.Errorf("init volume mount name = %q, want %q", init.VolumeMounts[0].Name, "creds")
	}
	if init.VolumeMounts[0].MountPath != "/tmp/creds" {
		t.Errorf("init volume mount path = %q, want %q", init.VolumeMounts[0].MountPath, "/tmp/creds")
	}
}

func TestBuildPodSpec_MainContainerClaudeSessionToken(t *testing.T) {
	p := defaultTestParams()
	pod := BuildPodSpec(p)

	if len(pod.Spec.Containers) != 1 {
		t.Fatalf("containers count = %d, want 1", len(pod.Spec.Containers))
	}
	agent := pod.Spec.Containers[0]

	var found bool
	for _, ev := range agent.Env {
		if ev.Name == "CLAUDE_CODE_OAUTH_TOKEN" {
			found = true
			if ev.ValueFrom == nil || ev.ValueFrom.SecretKeyRef == nil {
				t.Fatal("CLAUDE_CODE_OAUTH_TOKEN has no SecretKeyRef")
			}
			if ev.ValueFrom.SecretKeyRef.Name != p.Config.ClaudeSessionSecretRef {
				t.Errorf("secret name = %q, want %q", ev.ValueFrom.SecretKeyRef.Name, p.Config.ClaudeSessionSecretRef)
			}
			if ev.ValueFrom.SecretKeyRef.Key != "session-token" {
				t.Errorf("secret key = %q, want %q", ev.ValueFrom.SecretKeyRef.Key, "session-token")
			}
		}
	}
	if !found {
		t.Error("CLAUDE_CODE_OAUTH_TOKEN env var not found in main container")
	}
}

func TestBuildPodSpec_MainContainerInstanceEnvVars(t *testing.T) {
	p := defaultTestParams()
	pod := BuildPodSpec(p)
	agent := pod.Spec.Containers[0]

	envMap := make(map[string]string)
	for _, ev := range agent.Env {
		if ev.Value != "" {
			envMap[ev.Name] = ev.Value
		}
	}

	if envMap["AGENT_INSTANCE_ID"] != p.InstanceID {
		t.Errorf("AGENT_INSTANCE_ID = %q, want %q", envMap["AGENT_INSTANCE_ID"], p.InstanceID)
	}
	if envMap["AGENT_TEMPLATE"] != p.Template.Name {
		t.Errorf("AGENT_TEMPLATE = %q, want %q", envMap["AGENT_TEMPLATE"], p.Template.Name)
	}
}

func TestBuildPodSpec_MainContainerVolumeMount(t *testing.T) {
	pod := BuildPodSpec(defaultTestParams())
	agent := pod.Spec.Containers[0]

	if len(agent.VolumeMounts) != 1 {
		t.Fatalf("agent volume mounts = %d, want 1", len(agent.VolumeMounts))
	}
	vm := agent.VolumeMounts[0]
	if vm.Name != "creds" {
		t.Errorf("volume mount name = %q, want %q", vm.Name, "creds")
	}
	if vm.MountPath != "/tmp/creds" {
		t.Errorf("volume mount path = %q, want %q", vm.MountPath, "/tmp/creds")
	}
	if !vm.ReadOnly {
		t.Error("agent volume mount should be ReadOnly")
	}
}

func TestBuildPodSpec_Volumes(t *testing.T) {
	pod := BuildPodSpec(defaultTestParams())

	if len(pod.Spec.Volumes) != 1 {
		t.Fatalf("volumes count = %d, want 1", len(pod.Spec.Volumes))
	}
	vol := pod.Spec.Volumes[0]
	if vol.Name != "creds" {
		t.Errorf("volume name = %q, want %q", vol.Name, "creds")
	}
	if vol.VolumeSource.EmptyDir == nil {
		t.Error("volume source should be EmptyDir")
	}
}

func TestBuildPodSpec_EmptyCredentials(t *testing.T) {
	p := defaultTestParams()
	p.Credentials = map[string]string{}
	pod := BuildPodSpec(p)

	init := pod.Spec.InitContainers[0]
	if len(init.Env) != 0 {
		t.Errorf("init env vars = %d, want 0 for empty credentials", len(init.Env))
	}
	if !strings.HasPrefix(init.Command[2], "#!/bin/sh\n") {
		t.Error("write script should start with #!/bin/sh header even with no creds")
	}
}

func TestBuildPodSpec_MultipleCredentials(t *testing.T) {
	p := defaultTestParams()
	p.Credentials = map[string]string{
		"KEY_A": "val-a",
		"KEY_B": "val-b",
		"KEY_C": "val-c",
	}
	pod := BuildPodSpec(p)

	init := pod.Spec.InitContainers[0]
	if len(init.Env) != 3 {
		t.Errorf("init env vars = %d, want 3", len(init.Env))
	}

	envMap := make(map[string]string)
	for _, ev := range init.Env {
		envMap[ev.Name] = ev.Value
	}
	for k, v := range p.Credentials {
		if envMap[k] != v {
			t.Errorf("credential %q = %q, want %q", k, envMap[k], v)
		}
	}
}

func TestBuildPodSpec_ImageOverrides(t *testing.T) {
	p := defaultTestParams()
	p.ExportCredsImage = "custom-export-creds:dev"
	p.AgentImage = "custom-agent:dev"

	pod := BuildPodSpec(p)

	initImage := pod.Spec.InitContainers[0].Image
	if initImage != "custom-export-creds:dev" {
		t.Errorf("init container image = %q, want %q", initImage, "custom-export-creds:dev")
	}

	agentImage := pod.Spec.Containers[0].Image
	if agentImage != "custom-agent:dev" {
		t.Errorf("agent image = %q, want %q", agentImage, "custom-agent:dev")
	}
}

func TestBuildPodSpec_DefaultImages(t *testing.T) {
	p := defaultTestParams()
	// No image overrides — should use config and template values
	p.ExportCredsImage = ""
	p.AgentImage = ""

	pod := BuildPodSpec(p)

	if pod.Spec.InitContainers[0].Image != p.Config.ExportCredsImage {
		t.Errorf("init image should fall back to config, got %q", pod.Spec.InitContainers[0].Image)
	}
	if pod.Spec.Containers[0].Image != p.Template.DockerImage {
		t.Errorf("agent image should fall back to template, got %q", pod.Spec.Containers[0].Image)
	}
}

func TestBuildPodSpec_MainContainerCommand(t *testing.T) {
	pod := BuildPodSpec(defaultTestParams())
	agent := pod.Spec.Containers[0]

	if len(agent.Command) < 3 {
		t.Fatalf("agent command too short: %v", agent.Command)
	}
	cmd := agent.Command[2]
	if !strings.Contains(cmd, "/tmp/creds/env.sh") {
		t.Errorf("agent command should source creds, got: %q", cmd)
	}
	if !strings.Contains(cmd, "node /app/entrypoint.mjs") {
		t.Errorf("agent command should run entrypoint.mjs, got: %q", cmd)
	}
}

func TestBuildPodSpec_MainContainerMemoryLimits(t *testing.T) {
	pod := BuildPodSpec(defaultTestParams())
	agent := pod.Spec.Containers[0]

	memRequest := agent.Resources.Requests[corev1.ResourceMemory]
	if memRequest.String() != "512Mi" {
		t.Errorf("memory request = %q, want %q", memRequest.String(), "512Mi")
	}

	memLimit := agent.Resources.Limits[corev1.ResourceMemory]
	if memLimit.String() != "1Gi" {
		t.Errorf("memory limit = %q, want %q", memLimit.String(), "1Gi")
	}
}
