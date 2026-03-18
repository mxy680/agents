package orchestrator

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	podName := "agent-" + p.InstanceID[:8]

	exportCredsImage := p.Config.ExportCredsImage
	if p.ExportCredsImage != "" {
		exportCredsImage = p.ExportCredsImage
	}

	agentImage := p.Template.DockerImage
	if p.AgentImage != "" {
		agentImage = p.AgentImage
	}

	// Build credential env vars for init container
	credEnvVars := make([]corev1.EnvVar, 0, len(p.Credentials))
	for k, v := range p.Credentials {
		credEnvVars = append(credEnvVars, corev1.EnvVar{Name: k, Value: v})
	}

	// Build the shell script that writes env vars to /tmp/creds/env.sh
	writeCredsScript := "#!/bin/sh\n"
	for k := range p.Credentials {
		writeCredsScript += "echo 'export " + k + "='\\\"$" + k + "\\\"'' >> /tmp/creds/env.sh\n"
	}

	labels := map[string]string{
		"app":         "agent",
		"instance-id": p.InstanceID,
		"user-id":     p.UserID,
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
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
			},
		},
	}

	return pod
}
