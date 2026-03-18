package orchestrator

import (
	"encoding/json"
	"time"
)

// AgentTemplate represents a deployable agent type.
type AgentTemplate struct {
	ID                   string          `json:"id"`
	Name                 string          `json:"name"`
	DisplayName          string          `json:"display_name"`
	Description          string          `json:"description"`
	GitPath              string          `json:"git_path"`
	DockerImage          string          `json:"docker_image"`
	RequiredIntegrations []string        `json:"required_integrations"`
	DefaultConfig        json.RawMessage `json:"default_config"`
	Status               string          `json:"status"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

// AgentInstance represents a deployed agent for a specific user.
type AgentInstance struct {
	ID              string          `json:"id"`
	UserID          string          `json:"user_id"`
	TemplateID      string          `json:"template_id"`
	Status          string          `json:"status"`
	K8sPodName      string          `json:"k8s_pod_name,omitempty"`
	K8sNamespace    string          `json:"k8s_namespace"`
	ConfigOverrides json.RawMessage `json:"config_overrides"`
	ErrorMessage    string          `json:"error_message,omitempty"`
	StartedAt       *time.Time      `json:"started_at,omitempty"`
	StoppedAt       *time.Time      `json:"stopped_at,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// DeployRequest is the body for POST /agents/deploy.
type DeployRequest struct {
	TemplateID      string          `json:"template_id"`
	ConfigOverrides json.RawMessage `json:"config_overrides,omitempty"`
}

// Instance status constants.
const (
	StatusPending   = "pending"
	StatusCreating  = "creating"
	StatusRunning   = "running"
	StatusStopping  = "stopping"
	StatusStopped   = "stopped"
	StatusFailed    = "failed"
	StatusCompleted = "completed"
)
