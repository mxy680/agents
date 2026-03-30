package orchestrator

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
)

// templateColumns are the columns returned by template queries.
var templateColumns = []string{
	"id", "name", "display_name", "description", "git_path", "docker_image",
	"required_integrations", "default_config", "status", "created_at", "updated_at",
}

// instanceColumns are the columns returned by instance queries.
var instanceColumns = []string{
	"id", "user_id", "template_id", "status", "k8s_pod_name", "k8s_namespace",
	"config_overrides", "error_message", "started_at", "stopped_at", "created_at", "updated_at",
}

func newMockStore(t *testing.T) (*Store, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return NewStore(db), mock
}

var (
	fixedTime    = time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	defaultConfig = json.RawMessage(`{"key":"value"}`)
)

func addTemplateRow(rows *sqlmock.Rows, id, name string) *sqlmock.Rows {
	return rows.AddRow(
		id, name, name+" Display", "A description", "/path/to/"+name, "image:latest",
		pq.Array([]string{"google"}), defaultConfig, "active", fixedTime, fixedTime,
	)
}

func addInstanceRow(rows *sqlmock.Rows, id, userID, templateID, status string) *sqlmock.Rows {
	return rows.AddRow(
		id, userID, templateID, status,
		sql.NullString{Valid: false}, "agents",
		json.RawMessage(`{}`), sql.NullString{Valid: false},
		(*time.Time)(nil), (*time.Time)(nil),
		fixedTime, fixedTime,
	)
}

// ---- DB() ----

func TestStoreDB(t *testing.T) {
	store, _ := newMockStore(t)
	if store.DB() == nil {
		t.Error("DB() returned nil")
	}
}

// ---- ListTemplates ----

func TestListTemplates(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(mock sqlmock.Sqlmock)
		wantCount int
		wantErr   bool
	}{
		{
			name: "success with results",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(templateColumns)
				addTemplateRow(rows, "tmpl-1", "email-agent")
				addTemplateRow(rows, "tmpl-2", "calendar-agent")
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			wantCount: 2,
		},
		{
			name: "empty result",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(templateColumns)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			wantCount: 0,
		},
		{
			name: "query error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("connection refused"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, mock := newMockStore(t)
			tt.setup(mock)

			templates, err := store.ListTemplates(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("ListTemplates() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(templates) != tt.wantCount {
				t.Errorf("len(templates) = %d, want %d", len(templates), tt.wantCount)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// ---- GetTemplate ----

func TestGetTemplate(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			id:   "tmpl-1",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(templateColumns)
				addTemplateRow(rows, "tmpl-1", "email-agent")
				mock.ExpectQuery("SELECT").WithArgs("tmpl-1").WillReturnRows(rows)
			},
		},
		{
			name: "not found",
			id:   "tmpl-999",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(templateColumns)
				mock.ExpectQuery("SELECT").WithArgs("tmpl-999").WillReturnRows(rows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, mock := newMockStore(t)
			tt.setup(mock)

			tmpl, err := store.GetTemplate(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && tmpl.ID != tt.id {
				t.Errorf("template.ID = %q, want %q", tmpl.ID, tt.id)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// ---- UpsertTemplate ----

func TestUpsertTemplate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO agent_templates").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "db error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO agent_templates").
					WillReturnError(fmt.Errorf("duplicate key"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, mock := newMockStore(t)
			tt.setup(mock)

			tmpl := AgentTemplate{
				Name:                 "email-agent",
				DisplayName:          "Email Agent",
				Description:          "Handles email",
				GitPath:              "/agents/email",
				DockerImage:          "email-agent:latest",
				RequiredIntegrations: []string{"google"},
				DefaultConfig:        defaultConfig,
				Status:               "active",
			}

			err := store.UpsertTemplate(context.Background(), tmpl)
			if (err != nil) != tt.wantErr {
				t.Fatalf("UpsertTemplate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// ---- CreateInstance ----

func TestCreateInstance(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(instanceColumns)
				addInstanceRow(rows, "inst-1", "user-1", "tmpl-1", StatusPending)
				mock.ExpectQuery("INSERT INTO agent_instances").WillReturnRows(rows)
			},
		},
		{
			name: "db error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO agent_instances").
					WillReturnError(fmt.Errorf("fk violation"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, mock := newMockStore(t)
			tt.setup(mock)

			inst := AgentInstance{
				UserID:          "user-1",
				TemplateID:      "tmpl-1",
				K8sNamespace:    "agents",
				ConfigOverrides: json.RawMessage(`{}`),
			}

			created, err := store.CreateInstance(context.Background(), inst)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CreateInstance() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if created.ID != "inst-1" {
					t.Errorf("created.ID = %q, want %q", created.ID, "inst-1")
				}
				if created.Status != StatusPending {
					t.Errorf("created.Status = %q, want %q", created.Status, StatusPending)
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// ---- GetInstance ----

func TestGetInstance(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			id:   "inst-1",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(instanceColumns)
				addInstanceRow(rows, "inst-1", "user-1", "tmpl-1", StatusRunning)
				mock.ExpectQuery("SELECT").WithArgs("inst-1").WillReturnRows(rows)
			},
		},
		{
			name: "not found",
			id:   "inst-999",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(instanceColumns)
				mock.ExpectQuery("SELECT").WithArgs("inst-999").WillReturnRows(rows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, mock := newMockStore(t)
			tt.setup(mock)

			inst, err := store.GetInstance(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Fatalf("GetInstance() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && inst.ID != tt.id {
				t.Errorf("instance.ID = %q, want %q", inst.ID, tt.id)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// ---- ListInstances ----

func TestListInstances(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(mock sqlmock.Sqlmock)
		wantCount int
		wantErr   bool
	}{
		{
			name: "success with results",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(instanceColumns)
				addInstanceRow(rows, "inst-1", "user-1", "tmpl-1", StatusRunning)
				addInstanceRow(rows, "inst-2", "user-1", "tmpl-2", StatusPending)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			wantCount: 2,
		},
		{
			name: "empty result",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(instanceColumns)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			wantCount: 0,
		},
		{
			name: "query error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("timeout"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, mock := newMockStore(t)
			tt.setup(mock)

			instances, err := store.ListInstances(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("ListInstances() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(instances) != tt.wantCount {
				t.Errorf("len(instances) = %d, want %d", len(instances), tt.wantCount)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// ---- UpdateInstanceStatus ----

func TestUpdateInstanceStatus(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE agent_instances").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "db error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE agent_instances").
					WillReturnError(fmt.Errorf("connection lost"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, mock := newMockStore(t)
			tt.setup(mock)

			err := store.UpdateInstanceStatus(context.Background(), "inst-1", StatusRunning, "pod-abc", "")
			if (err != nil) != tt.wantErr {
				t.Fatalf("UpdateInstanceStatus() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// ---- DeleteInstance ----

func TestDeleteInstance(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM agent_instances").
					WithArgs("inst-1").
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},
		{
			name: "not in terminal state (0 rows affected)",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM agent_instances").
					WithArgs("inst-1").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "db error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM agent_instances").
					WithArgs("inst-1").
					WillReturnError(fmt.Errorf("permission denied"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, mock := newMockStore(t)
			tt.setup(mock)

			err := store.DeleteInstance(context.Background(), "inst-1")
			if (err != nil) != tt.wantErr {
				t.Fatalf("DeleteInstance() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// ---- ListRunningInstances ----

func TestListRunningInstances(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(mock sqlmock.Sqlmock)
		wantCount int
		wantErr   bool
	}{
		{
			name: "success with results",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(instanceColumns)
				addInstanceRow(rows, "inst-1", "user-1", "tmpl-1", StatusRunning)
				addInstanceRow(rows, "inst-2", "user-2", "tmpl-1", StatusPending)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			wantCount: 2,
		},
		{
			name: "empty result",
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(instanceColumns)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			wantCount: 0,
		},
		{
			name: "query error",
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").WillReturnError(fmt.Errorf("timeout"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, mock := newMockStore(t)
			tt.setup(mock)

			instances, err := store.ListRunningInstances(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("ListRunningInstances() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(instances) != tt.wantCount {
				t.Errorf("len(instances) = %d, want %d", len(instances), tt.wantCount)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}

// ---- CheckIntegrations ----

func TestCheckIntegrations(t *testing.T) {
	tests := []struct {
		name        string
		required    []string
		setup       func(mock sqlmock.Sqlmock)
		wantMissing []string
		wantErr     bool
	}{
		{
			name:     "none required",
			required: nil,
			setup:    func(mock sqlmock.Sqlmock) {},
		},
		{
			name:     "all connected",
			required: []string{"google", "github"},
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"provider"}).
					AddRow("google").
					AddRow("github")
				mock.ExpectQuery("SELECT DISTINCT provider").WithArgs(pq.Array([]string{"google", "github"})).WillReturnRows(rows)
			},
			wantMissing: nil,
		},
		{
			name:     "some missing",
			required: []string{"google", "github", "instagram"},
			setup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"provider"}).
					AddRow("google")
				mock.ExpectQuery("SELECT DISTINCT provider").WithArgs(pq.Array([]string{"google", "github", "instagram"})).WillReturnRows(rows)
			},
			wantMissing: []string{"github", "instagram"},
		},
		{
			name:     "query error",
			required: []string{"google"},
			setup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT DISTINCT provider").WillReturnError(fmt.Errorf("timeout"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, mock := newMockStore(t)
			tt.setup(mock)

			missing, err := store.CheckIntegrations(context.Background(), tt.required)
			if (err != nil) != tt.wantErr {
				t.Fatalf("CheckIntegrations() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if len(missing) != len(tt.wantMissing) {
					t.Errorf("missing = %v, want %v", missing, tt.wantMissing)
				} else {
					missingSet := make(map[string]bool)
					for _, m := range missing {
						missingSet[m] = true
					}
					for _, want := range tt.wantMissing {
						if !missingSet[want] {
							t.Errorf("expected %q in missing, got %v", want, missing)
						}
					}
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("unfulfilled expectations: %v", err)
			}
		})
	}
}
