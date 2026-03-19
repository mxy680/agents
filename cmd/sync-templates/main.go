package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/emdash-projects/agents/internal/orchestrator"
	_ "github.com/lib/pq"
	"gopkg.in/yaml.v3"
)

// templateYAML mirrors the template.yaml file format.
type templateYAML struct {
	Name                 string         `yaml:"name"`
	DisplayName          string         `yaml:"display_name"`
	Description          string         `yaml:"description"`
	GitPath              string         `yaml:"git_path"`
	DockerImage          string         `yaml:"docker_image"`
	RequiredIntegrations []string       `yaml:"required_integrations"`
	DefaultConfig        map[string]any `yaml:"default_config"`
	Status               string         `yaml:"status"`
}

// jobsYAML mirrors the jobs/jobs.yaml file format.
type jobsYAML struct {
	Jobs []jobYAML `yaml:"jobs"`
}

// jobYAML mirrors a single job entry in jobs/jobs.yaml.
type jobYAML struct {
	Slug           string `yaml:"slug"`
	DisplayName    string `yaml:"display_name"`
	Description    string `yaml:"description"`
	Schedule       string `yaml:"schedule"`
	PromptFile     string `yaml:"prompt_file"`
	TimeoutMinutes int    `yaml:"timeout_minutes"`
}

func main() {
	dbURL := os.Getenv("SUPABASE_DB_URL")
	if dbURL == "" {
		log.Fatal("SUPABASE_DB_URL not set")
	}

	agentsDir := "agents"
	if len(os.Args) > 1 {
		agentsDir = os.Args[1]
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer db.Close()

	store := orchestrator.NewStore(db)
	ctx := context.Background()

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		log.Fatalf("read agents dir: %v", err)
	}

	synced := 0
	jobsSynced := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		yamlPath := filepath.Join(agentsDir, entry.Name(), "template.yaml")
		data, err := os.ReadFile(yamlPath)
		if err != nil {
			log.Printf("skip %s: %v", entry.Name(), err)
			continue
		}

		var tmpl templateYAML
		if err := yaml.Unmarshal(data, &tmpl); err != nil {
			log.Printf("skip %s: invalid yaml: %v", entry.Name(), err)
			continue
		}

		defaultCfg, _ := json.Marshal(tmpl.DefaultConfig)

		t := orchestrator.AgentTemplate{
			Name:                 tmpl.Name,
			DisplayName:          tmpl.DisplayName,
			Description:          tmpl.Description,
			GitPath:              tmpl.GitPath,
			DockerImage:          tmpl.DockerImage,
			RequiredIntegrations: tmpl.RequiredIntegrations,
			DefaultConfig:        defaultCfg,
			Status:               tmpl.Status,
		}

		if err := store.UpsertTemplate(ctx, t); err != nil {
			log.Printf("error syncing %s: %v", tmpl.Name, err)
			continue
		}

		fmt.Printf("synced: %s\n", tmpl.Name)
		synced++

		// Sync job definitions for this template.
		n, err := syncJobs(ctx, db, agentsDir, entry.Name(), tmpl.Name)
		if err != nil {
			log.Printf("error syncing jobs for %s: %v", tmpl.Name, err)
		}
		jobsSynced += n
	}

	fmt.Printf("\n%d template(s) synced\n", synced)
	fmt.Printf("%d job definition(s) synced\n", jobsSynced)
}

// syncJobs reads jobs/jobs.yaml for an agent directory, upserts job definitions,
// and disables any existing job definitions for that template not in the manifest.
// Returns the number of job definitions synced.
func syncJobs(ctx context.Context, db *sql.DB, agentsDir, agentDirName, templateName string) (int, error) {
	jobsYAMLPath := filepath.Join(agentsDir, agentDirName, "jobs", "jobs.yaml")
	data, err := os.ReadFile(jobsYAMLPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read jobs.yaml: %w", err)
	}

	var manifest jobsYAML
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return 0, fmt.Errorf("parse jobs.yaml: %w", err)
	}

	if len(manifest.Jobs) == 0 {
		return 0, nil
	}

	// Look up the template_id by name.
	var templateID string
	err = db.QueryRowContext(ctx,
		`SELECT id FROM agent_templates WHERE name = $1`, templateName).
		Scan(&templateID)
	if err != nil {
		return 0, fmt.Errorf("look up template %q: %w", templateName, err)
	}

	// Collect the slugs present in this manifest for later cleanup.
	manifestSlugs := make([]string, 0, len(manifest.Jobs))

	synced := 0
	for _, job := range manifest.Jobs {
		promptPath := filepath.Join(agentsDir, agentDirName, "jobs", job.PromptFile)
		promptBytes, err := os.ReadFile(promptPath)
		if err != nil {
			log.Printf("  skip job %s/%s: read prompt file %s: %v", templateName, job.Slug, job.PromptFile, err)
			continue
		}
		prompt := string(promptBytes)

		timeoutMinutes := job.TimeoutMinutes
		if timeoutMinutes <= 0 {
			timeoutMinutes = 10
		}

		_, err = db.ExecContext(ctx,
			`INSERT INTO job_definitions
			   (template_id, slug, display_name, description, schedule, prompt, timeout_minutes, enabled)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, true)
			 ON CONFLICT (template_id, slug) DO UPDATE SET
			   display_name     = EXCLUDED.display_name,
			   description      = EXCLUDED.description,
			   schedule         = EXCLUDED.schedule,
			   prompt           = EXCLUDED.prompt,
			   timeout_minutes  = EXCLUDED.timeout_minutes,
			   enabled          = true,
			   updated_at       = now()`,
			templateID, job.Slug, job.DisplayName, job.Description,
			job.Schedule, prompt, timeoutMinutes)
		if err != nil {
			log.Printf("  error upserting job %s/%s: %v", templateName, job.Slug, err)
			continue
		}

		fmt.Printf("  synced job: %s/%s\n", templateName, job.Slug)
		manifestSlugs = append(manifestSlugs, job.Slug)
		synced++
	}

	// Disable job definitions for this template that are no longer in the manifest.
	if len(manifestSlugs) > 0 {
		slugArray := make([]interface{}, len(manifestSlugs))
		for i, s := range manifestSlugs {
			slugArray[i] = s
		}
		placeholders := makePlaceholders(2, len(manifestSlugs))
		args := append([]interface{}{templateID}, slugArray...)
		_, err = db.ExecContext(ctx,
			`UPDATE job_definitions
			 SET enabled = false, updated_at = now()
			 WHERE template_id = $1 AND slug NOT IN (`+placeholders+`)`,
			args...)
		if err != nil {
			log.Printf("  error disabling stale jobs for %s: %v", templateName, err)
		}
	} else {
		// No jobs in manifest — disable all for this template.
		_, err = db.ExecContext(ctx,
			`UPDATE job_definitions
			 SET enabled = false, updated_at = now()
			 WHERE template_id = $1`,
			templateID)
		if err != nil {
			log.Printf("  error disabling all jobs for %s: %v", templateName, err)
		}
	}

	return synced, nil
}

// makePlaceholders generates a comma-separated list of $N placeholders
// starting at offset+1. E.g. makePlaceholders(2, 3) → "$2,$3,$4"
func makePlaceholders(offset, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("$%d", offset+i)
	}
	return result
}
