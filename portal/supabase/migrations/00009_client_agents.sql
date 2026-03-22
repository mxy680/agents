-- Migration: client_agents join table
-- Tracks which agent templates are assigned to which clients (many-to-many)

CREATE TABLE IF NOT EXISTS client_agents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
  template_id UUID NOT NULL REFERENCES agent_templates(id) ON DELETE CASCADE,
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(client_id, template_id)
);

CREATE INDEX IF NOT EXISTS idx_client_agents_client ON client_agents(client_id);
CREATE INDEX IF NOT EXISTS idx_client_agents_template ON client_agents(template_id);
