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
	}

	fmt.Printf("\n%d template(s) synced\n", synced)
}
