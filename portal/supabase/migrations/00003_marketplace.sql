-- Track which agents a user has acquired
CREATE TABLE IF NOT EXISTS user_agents (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  template_id UUID NOT NULL REFERENCES agent_templates(id) ON DELETE CASCADE,
  acquired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(user_id, template_id)
);

ALTER TABLE user_agents ENABLE ROW LEVEL SECURITY;

CREATE POLICY "users_select_own_agents" ON user_agents
  FOR SELECT TO authenticated USING (auth.uid() = user_id);
CREATE POLICY "users_insert_own_agents" ON user_agents
  FOR INSERT TO authenticated WITH CHECK (auth.uid() = user_id);
CREATE POLICY "users_delete_own_agents" ON user_agents
  FOR DELETE TO authenticated USING (auth.uid() = user_id);

CREATE INDEX idx_user_agents_user_id ON user_agents(user_id);

-- Extend agent_templates with marketplace metadata
ALTER TABLE agent_templates ADD COLUMN IF NOT EXISTS category TEXT NOT NULL DEFAULT 'general';
ALTER TABLE agent_templates ADD COLUMN IF NOT EXISTS icon_url TEXT;
ALTER TABLE agent_templates ADD COLUMN IF NOT EXISTS tags TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE agent_templates ADD COLUMN IF NOT EXISTS acquisition_count INT NOT NULL DEFAULT 0;
