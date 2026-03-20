-- Conversations (per user, per agent)
CREATE TABLE IF NOT EXISTS conversations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  agent_name TEXT NOT NULL,
  title TEXT,
  starred BOOLEAN NOT NULL DEFAULT FALSE,
  session_id TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Conversation messages
CREATE TABLE IF NOT EXISTS conversation_messages (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES auth.users(id) ON DELETE CASCADE,
  role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
  blocks JSONB NOT NULL DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- RLS
ALTER TABLE conversations ENABLE ROW LEVEL SECURITY;
CREATE POLICY "conversations_select" ON conversations
  FOR SELECT TO authenticated USING (auth.uid() = user_id);
CREATE POLICY "conversations_insert" ON conversations
  FOR INSERT TO authenticated WITH CHECK (auth.uid() = user_id);
CREATE POLICY "conversations_update" ON conversations
  FOR UPDATE TO authenticated USING (auth.uid() = user_id);
CREATE POLICY "conversations_delete" ON conversations
  FOR DELETE TO authenticated USING (auth.uid() = user_id);

ALTER TABLE conversation_messages ENABLE ROW LEVEL SECURITY;
CREATE POLICY "conversation_messages_select" ON conversation_messages
  FOR SELECT TO authenticated USING (auth.uid() = user_id);
CREATE POLICY "conversation_messages_insert" ON conversation_messages
  FOR INSERT TO authenticated WITH CHECK (auth.uid() = user_id);
CREATE POLICY "conversation_messages_insert_service" ON conversation_messages
  FOR INSERT TO service_role WITH CHECK (TRUE);
CREATE POLICY "conversation_messages_update_service" ON conversation_messages
  FOR UPDATE TO service_role USING (TRUE);

-- Indexes
CREATE INDEX idx_conversations_user_agent ON conversations(user_id, agent_name, updated_at DESC);
CREATE INDEX idx_conversations_starred ON conversations(user_id, starred) WHERE starred = TRUE;
CREATE INDEX idx_conversation_messages_conversation ON conversation_messages(conversation_id, created_at ASC);
