package orchestrator

import (
	"fmt"
	"os"
	"regexp"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// validCredKeyRe matches safe environment variable names only (e.g. GOOGLE_ACCESS_TOKEN).
var validCredKeyRe = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

// PodSpecParams contains all inputs needed to build an agent pod spec.
type PodSpecParams struct {
	InstanceID  string
	UserID      string
	Template    AgentTemplate
	Namespace   string
	Credentials map[string]string
	Config      Config

	// Override images (for testing)
	ExportCredsImage string
	AgentImage       string
}

// BuildPodSpec creates a Kubernetes Pod specification for an agent deployment.
func BuildPodSpec(p PodSpecParams) *corev1.Pod {
	// Guard against short InstanceID to prevent slice panic.
	id := p.InstanceID
	if len(id) > 8 {
		id = id[:8]
	}
	podName := "agent-" + id

	exportCredsImage := p.Config.ExportCredsImage
	if p.ExportCredsImage != "" {
		exportCredsImage = p.ExportCredsImage
	}

	agentImage := p.Template.DockerImage
	if p.AgentImage != "" {
		agentImage = p.AgentImage
	}

	// Sort credential keys for deterministic output and to simplify auditing.
	credKeys := make([]string, 0, len(p.Credentials))
	for k := range p.Credentials {
		if validCredKeyRe.MatchString(k) {
			credKeys = append(credKeys, k)
		} else {
			fmt.Fprintf(os.Stderr, "WARNING: skipping credential key with unsafe name: %q\n", k)
		}
	}
	sort.Strings(credKeys)

	// Build credential env vars for init container using only validated keys.
	credEnvVars := make([]corev1.EnvVar, 0, len(credKeys))
	for _, k := range credKeys {
		credEnvVars = append(credEnvVars, corev1.EnvVar{Name: k, Value: p.Credentials[k]})
	}

	// Build the shell script that writes env vars to /tmp/creds/env.sh.
	// Keys are pre-validated against validCredKeyRe so no injection is possible.
	writeCredsScript := "#!/bin/sh\n"
	for _, k := range credKeys {
		writeCredsScript += "echo 'export " + k + "='\\\"$" + k + "\\\"'' >> /tmp/creds/env.sh\n"
	}

	labels := map[string]string{
		"app":         "agent",
		"instance-id": p.InstanceID,
		"user-id":     p.UserID,
	}

	// Security contexts.
	falseVal := false
	trueVal := true
	runAsUser := int64(1000)

	// Init container runs as root to write creds but must not escalate privileges.
	initSecCtx := &corev1.SecurityContext{
		RunAsNonRoot:             &falseVal,
		AllowPrivilegeEscalation: &falseVal,
	}

	// Main agent container runs as non-root uid 1000 (the 'agent' user from the Dockerfile).
	agentSecCtx := &corev1.SecurityContext{
		RunAsNonRoot:             &trueVal,
		AllowPrivilegeEscalation: &falseVal,
		RunAsUser:                &runAsUser,
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: p.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			InitContainers: []corev1.Container{
				{
					Name:            "resolve-creds",
					Image:           exportCredsImage,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         []string{"/bin/sh", "-c", writeCredsScript},
					Env:             credEnvVars,
					SecurityContext: initSecCtx,
					VolumeMounts: []corev1.VolumeMount{
						{Name: "creds", MountPath: "/tmp/creds"},
					},
					Resources: corev1.ResourceRequirements{
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("100m"),
							corev1.ResourceMemory: resource.MustParse("64Mi"),
						},
					},
				},
			},
			Containers: []corev1.Container{
				{
					Name:            "agent",
					Image:           agentImage,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Command:         []string{"/bin/sh", "-c", "if [ -f /tmp/creds/env.sh ]; then . /tmp/creds/env.sh; fi && node /app/entrypoint.mjs"},
					SecurityContext: agentSecCtx,
					Env: []corev1.EnvVar{
						{
							Name: "CLAUDE_CODE_OAUTH_TOKEN",
							ValueFrom: &corev1.EnvVarSource{
								SecretKeyRef: &corev1.SecretKeySelector{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: p.Config.ClaudeSessionSecretRef,
									},
									Key: "session-token",
								},
							},
						},
						{Name: "AGENT_INSTANCE_ID", Value: p.InstanceID},
						{Name: "AGENT_TEMPLATE", Value: p.Template.Name},
					},
					VolumeMounts: []corev1.VolumeMount{
						{Name: "creds", MountPath: "/tmp/creds", ReadOnly: true},
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("250m"),
							corev1.ResourceMemory: resource.MustParse("512Mi"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceCPU:    resource.MustParse("1"),
							corev1.ResourceMemory: resource.MustParse("1Gi"),
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "creds",
					VolumeSource: corev1.VolumeSource{
						// Use Memory medium so credentials are never flushed to disk.
						EmptyDir: &corev1.EmptyDirVolumeSource{
							Medium: corev1.StorageMediumMemory,
						},
					},
				},
			},
		},
	}

	return pod
}
