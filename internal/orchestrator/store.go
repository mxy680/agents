package orchestrator

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

// Store provides CRUD operations for agent templates and instances.
type Store struct {
	db *sql.DB
}

// NewStore creates a new Store backed by the given database.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// DB returns the underlying database connection (implements tokenbridge.DB).
func (s *Store) DB() *sql.DB {
	return s.db
}

// ListTemplates returns all active templates.
func (s *Store) ListTemplates(ctx context.Context) ([]AgentTemplate, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, display_name, description, git_path, docker_image,
		        required_integrations, default_config, status, created_at, updated_at
		 FROM agent_templates
		 WHERE status = 'active'
		 ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("query templates: %w", err)
	}
	defer rows.Close()

	var templates []AgentTemplate
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

// GetTemplate returns a template by ID.
func (s *Store) GetTemplate(ctx context.Context, id string) (AgentTemplate, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, display_name, description, git_path, docker_image,
		        required_integrations, default_config, status, created_at, updated_at
		 FROM agent_templates
		 WHERE id = $1`, id)

	return scanTemplateRow(row)
}

// UpsertTemplate inserts or updates a template by name.
func (s *Store) UpsertTemplate(ctx context.Context, t AgentTemplate) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO agent_templates (name, display_name, description, git_path, docker_image,
		                              required_integrations, default_config, status)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 ON CONFLICT (name) DO UPDATE SET
		   display_name = EXCLUDED.display_name,
		   description = EXCLUDED.description,
		   git_path = EXCLUDED.git_path,
		   docker_image = EXCLUDED.docker_image,
		   required_integrations = EXCLUDED.required_integrations,
		   default_config = EXCLUDED.default_config,
		   status = EXCLUDED.status,
		   updated_at = now()`,
		t.Name, t.DisplayName, t.Description, t.GitPath, t.DockerImage,
		pq.Array(t.RequiredIntegrations), t.DefaultConfig, t.Status)
	if err != nil {
		return fmt.Errorf("upsert template: %w", err)
	}
	return nil
}

// CreateInstance inserts a new agent instance.
func (s *Store) CreateInstance(ctx context.Context, inst AgentInstance) (AgentInstance, error) {
	row := s.db.QueryRowContext(ctx,
		`INSERT INTO agent_instances (user_id, template_id, status, k8s_namespace, config_overrides)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, user_id, template_id, status, k8s_pod_name, k8s_namespace,
		           config_overrides, error_message, started_at, stopped_at, created_at, updated_at`,
		inst.UserID, inst.TemplateID, StatusPending, inst.K8sNamespace, inst.ConfigOverrides)

	created, err := scanInstanceRow(row)
	if err != nil {
		return AgentInstance{}, fmt.Errorf("create instance: %w", err)
	}
	return created, nil
}

// GetInstance returns an instance by ID.
func (s *Store) GetInstance(ctx context.Context, id string) (AgentInstance, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, user_id, template_id, status, k8s_pod_name, k8s_namespace,
		        config_overrides, error_message, started_at, stopped_at, created_at, updated_at
		 FROM agent_instances
		 WHERE id = $1`, id)

	return scanInstanceRow(row)
}

// ListInstances returns all instances for a user.
func (s *Store) ListInstances(ctx context.Context, userID string) ([]AgentInstance, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, template_id, status, k8s_pod_name, k8s_namespace,
		        config_overrides, error_message, started_at, stopped_at, created_at, updated_at
		 FROM agent_instances
		 WHERE user_id = $1
		 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("query instances: %w", err)
	}
	defer rows.Close()

	var instances []AgentInstance
	for rows.Next() {
		inst, err := scanInstance(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

// UpdateInstanceStatus updates the status and optional fields of an instance.
func (s *Store) UpdateInstanceStatus(ctx context.Context, id, status string, podName string, errMsg string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE agent_instances
		 SET status = $2,
		     k8s_pod_name = COALESCE(NULLIF($3, ''), k8s_pod_name),
		     error_message = CASE WHEN $4 = '' THEN error_message ELSE $4 END,
		     started_at = CASE WHEN $2 = 'running' AND started_at IS NULL THEN now() ELSE started_at END,
		     stopped_at = CASE WHEN $2 IN ('stopped', 'failed', 'completed') THEN now() ELSE stopped_at END,
		     updated_at = now()
		 WHERE id = $1`, id, status, podName, errMsg)
	if err != nil {
		return fmt.Errorf("update instance status: %w", err)
	}
	return nil
}

// DeleteInstance deletes a stopped/failed/completed instance.
func (s *Store) DeleteInstance(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx,
		`DELETE FROM agent_instances
		 WHERE id = $1 AND status IN ('stopped', 'failed', 'completed')`, id)
	if err != nil {
		return fmt.Errorf("delete instance: %w", err)
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("instance not found or not in terminal state")
	}
	return nil
}

// ListRunningInstances returns all instances in non-terminal states (for reconciliation).
func (s *Store) ListRunningInstances(ctx context.Context) ([]AgentInstance, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, user_id, template_id, status, k8s_pod_name, k8s_namespace,
		        config_overrides, error_message, started_at, stopped_at, created_at, updated_at
		 FROM agent_instances
		 WHERE status IN ('pending', 'creating', 'running', 'stopping')
		 ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("query running instances: %w", err)
	}
	defer rows.Close()

	var instances []AgentInstance
	for rows.Next() {
		inst, err := scanInstance(rows)
		if err != nil {
			return nil, err
		}
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}

// CheckUserIntegrations verifies a user has all required integrations connected.
// Returns the slice of missing provider names (nil if all connected).
func (s *Store) CheckUserIntegrations(ctx context.Context, userID string, required []string) ([]string, error) {
	if len(required) == 0 {
		return nil, nil
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT DISTINCT provider
		 FROM user_integrations
		 WHERE user_id = $1 AND status = 'active' AND provider = ANY($2)`,
		userID, pq.Array(required))
	if err != nil {
		return nil, fmt.Errorf("check integrations: %w", err)
	}
	defer rows.Close()

	connected := make(map[string]bool)
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		connected[p] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var missing []string
	for _, r := range required {
		if !connected[r] {
			missing = append(missing, r)
		}
	}
	return missing, nil
}

// scanner is implemented by both *sql.Rows and *sql.Row.
type scanner interface {
	Scan(dest ...any) error
}

func scanTemplate(s scanner) (AgentTemplate, error) {
	var t AgentTemplate
	var reqIntegrations []string
	err := s.Scan(&t.ID, &t.Name, &t.DisplayName, &t.Description, &t.GitPath, &t.DockerImage,
		pq.Array(&reqIntegrations), &t.DefaultConfig, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return AgentTemplate{}, fmt.Errorf("scan template: %w", err)
	}
	t.RequiredIntegrations = reqIntegrations
	return t, nil
}

func scanTemplateRow(row *sql.Row) (AgentTemplate, error) {
	var t AgentTemplate
	var reqIntegrations []string
	err := row.Scan(&t.ID, &t.Name, &t.DisplayName, &t.Description, &t.GitPath, &t.DockerImage,
		pq.Array(&reqIntegrations), &t.DefaultConfig, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return AgentTemplate{}, fmt.Errorf("scan template: %w", err)
	}
	t.RequiredIntegrations = reqIntegrations
	return t, nil
}

func scanInstance(s scanner) (AgentInstance, error) {
	var inst AgentInstance
	var podName, errMsg sql.NullString
	err := s.Scan(&inst.ID, &inst.UserID, &inst.TemplateID, &inst.Status,
		&podName, &inst.K8sNamespace, &inst.ConfigOverrides, &errMsg,
		&inst.StartedAt, &inst.StoppedAt, &inst.CreatedAt, &inst.UpdatedAt)
	if err != nil {
		return AgentInstance{}, fmt.Errorf("scan instance: %w", err)
	}
	inst.K8sPodName = podName.String
	inst.ErrorMessage = errMsg.String
	return inst, nil
}

func scanInstanceRow(row *sql.Row) (AgentInstance, error) {
	var inst AgentInstance
	var podName, errMsg sql.NullString
	err := row.Scan(&inst.ID, &inst.UserID, &inst.TemplateID, &inst.Status,
		&podName, &inst.K8sNamespace, &inst.ConfigOverrides, &errMsg,
		&inst.StartedAt, &inst.StoppedAt, &inst.CreatedAt, &inst.UpdatedAt)
	if err != nil {
		return AgentInstance{}, fmt.Errorf("scan instance: %w", err)
	}
	inst.K8sPodName = podName.String
	inst.ErrorMessage = errMsg.String
	return inst, nil
}
