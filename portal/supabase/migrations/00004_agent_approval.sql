-- Add approval status to user_agents
ALTER TABLE user_agents ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'pending'
  CHECK (status IN ('pending', 'approved', 'rejected'));
ALTER TABLE user_agents ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMPTZ;
ALTER TABLE user_agents ADD COLUMN IF NOT EXISTS reviewer_note TEXT;
