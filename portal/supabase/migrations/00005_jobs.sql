-- Job definitions synced from git (per agent template)
CREATE TABLE IF NOT EXISTS job_definitions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  template_id UUID NOT NULL REFERENCES agent_templates(id) ON DELETE CASCADE,
  slug TEXT NOT NULL,
  display_name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  schedule TEXT NOT NULL,
  prompt TEXT NOT NULL,
  timeout_minutes INT NOT NULL DEFAULT 10,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE(template_id, slug)
);

-- Job run history (per user, per job definition)
CREATE TABLE IF NOT EXISTS job_runs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  job_definition_id UUID NOT NULL REFERENCES job_definitions(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  status TEXT NOT NULL DEFAULT 'pending'
    CHECK (status IN ('pending', 'running', 'completed', 'failed', 'timed_out')),
  output_markdown TEXT,
  error_message TEXT,
  started_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  duration_ms INT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- RLS
ALTER TABLE job_definitions ENABLE ROW LEVEL SECURITY;
CREATE POLICY "job_definitions_select" ON job_definitions
  FOR SELECT TO authenticated USING (TRUE);

ALTER TABLE job_runs ENABLE ROW LEVEL SECURITY;
CREATE POLICY "job_runs_select" ON job_runs
  FOR SELECT TO authenticated USING (auth.uid() = user_id);
CREATE POLICY "job_runs_insert_service" ON job_runs
  FOR INSERT TO service_role WITH CHECK (TRUE);
CREATE POLICY "job_runs_update_service" ON job_runs
  FOR UPDATE TO service_role USING (TRUE);

-- Indexes
CREATE INDEX idx_job_definitions_template ON job_definitions(template_id);
CREATE INDEX idx_job_runs_user ON job_runs(user_id);
CREATE INDEX idx_job_runs_definition ON job_runs(job_definition_id);
CREATE INDEX idx_job_runs_status ON job_runs(status);
CREATE INDEX idx_job_runs_created ON job_runs(created_at DESC);
