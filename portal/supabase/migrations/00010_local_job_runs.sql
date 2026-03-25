CREATE TABLE IF NOT EXISTS local_job_runs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  agent_name TEXT NOT NULL,
  job_slug TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'pending',
  started_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  log TEXT DEFAULT '',
  deliverables JSONB DEFAULT '{}',
  created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_local_job_runs_status ON local_job_runs(status);
CREATE INDEX idx_local_job_runs_agent ON local_job_runs(agent_name);
