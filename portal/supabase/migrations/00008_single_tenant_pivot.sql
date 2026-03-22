-- Single-tenant pivot: introduce clients table, link agent_instances/conversations/job_runs
-- to clients, and remove marketplace tables/columns no longer needed in an internal tool.

-- 1. Create clients table
CREATE TABLE IF NOT EXISTS clients (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  email text,
  notes text,
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- 2. Add optional client_id to agent_instances
ALTER TABLE agent_instances ADD COLUMN IF NOT EXISTS client_id uuid REFERENCES clients(id);

-- 3. Add optional client_id to conversations
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS client_id uuid REFERENCES clients(id);

-- 4. Add optional client_id to job_runs
ALTER TABLE job_runs ADD COLUMN IF NOT EXISTS client_id uuid REFERENCES clients(id);

-- 5. Drop user_agents (self-service acquisition no longer needed)
DROP TABLE IF EXISTS user_agents CASCADE;

-- 6. Drop increment_acquisition_count function
DROP FUNCTION IF EXISTS increment_acquisition_count;

-- 7. Remove acquisition_count from agent_templates
ALTER TABLE agent_templates DROP COLUMN IF EXISTS acquisition_count;

-- 8. Remove marketplace columns from agent_templates
ALTER TABLE agent_templates DROP COLUMN IF EXISTS category;
ALTER TABLE agent_templates DROP COLUMN IF EXISTS icon_url;
ALTER TABLE agent_templates DROP COLUMN IF EXISTS tags;

-- 9. Indexes
CREATE INDEX IF NOT EXISTS idx_clients_active ON clients(active);
CREATE INDEX IF NOT EXISTS idx_agent_instances_client ON agent_instances(client_id);
CREATE INDEX IF NOT EXISTS idx_conversations_client ON conversations(client_id);
CREATE INDEX IF NOT EXISTS idx_job_runs_client ON job_runs(client_id);
