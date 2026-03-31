import { Pool } from 'pg'
import { execFile } from 'child_process'
import { promisify } from 'util'

const execFileAsync = promisify(execFile)

const CONTAINER_NAME = 'ade-postgres'
const PG_PORT = 5434
const PG_USER = 'ade'
const PG_PASSWORD = 'ade'
const PG_DB = 'ade'

const CONNECTION_STRING = `postgresql://${PG_USER}:${PG_PASSWORD}@localhost:${PG_PORT}/${PG_DB}`

let pool: Pool | null = null

async function ensurePostgresRunning(): Promise<void> {
  try {
    // Check if container exists and is running
    const { stdout } = await execFileAsync('docker', [
      'inspect', '-f', '{{.State.Running}}', CONTAINER_NAME
    ])
    if (stdout.trim() === 'true') return

    // Container exists but stopped — start it
    await execFileAsync('docker', ['start', CONTAINER_NAME])
    await waitForPostgres()
    return
  } catch {
    // Container doesn't exist — create it
  }

  await execFileAsync('docker', [
    'run', '-d',
    '--name', CONTAINER_NAME,
    '-e', `POSTGRES_USER=${PG_USER}`,
    '-e', `POSTGRES_PASSWORD=${PG_PASSWORD}`,
    '-e', `POSTGRES_DB=${PG_DB}`,
    '-p', `${PG_PORT}:5432`,
    '--restart', 'unless-stopped',
    'postgres:17-alpine'
  ])

  await waitForPostgres()
}

async function waitForPostgres(): Promise<void> {
  for (let i = 0; i < 30; i++) {
    try {
      const testPool = new Pool({ connectionString: CONNECTION_STRING })
      await testPool.query('SELECT 1')
      await testPool.end()
      return
    } catch {
      await new Promise((r) => setTimeout(r, 1000))
    }
  }
  throw new Error('Postgres failed to start within 30 seconds')
}

async function getPool(): Promise<Pool> {
  if (!pool) {
    await ensurePostgresRunning()
    pool = new Pool({ connectionString: CONNECTION_STRING })
    await migrate()
  }
  return pool
}

async function migrate(): Promise<void> {
  const p = pool!
  await p.query(`
    CREATE TABLE IF NOT EXISTS messages (
      id TEXT PRIMARY KEY,
      project_id TEXT NOT NULL,
      role TEXT NOT NULL,
      content TEXT NOT NULL,
      timestamp TEXT NOT NULL,
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
    CREATE INDEX IF NOT EXISTS idx_messages_project ON messages(project_id);

    CREATE TABLE IF NOT EXISTS minions (
      id TEXT PRIMARY KEY,
      project_id TEXT NOT NULL,
      name TEXT NOT NULL,
      task TEXT NOT NULL,
      branch TEXT NOT NULL,
      worktree_path TEXT NOT NULL,
      project_path TEXT NOT NULL,
      status TEXT NOT NULL DEFAULT 'spawning',
      claude_session_id TEXT,
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
    CREATE INDEX IF NOT EXISTS idx_minions_project ON minions(project_id);

    CREATE TABLE IF NOT EXISTS minion_logs (
      id TEXT PRIMARY KEY,
      minion_id TEXT NOT NULL,
      project_id TEXT NOT NULL,
      role TEXT NOT NULL,
      content TEXT NOT NULL,
      timestamp TEXT NOT NULL,
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
    CREATE INDEX IF NOT EXISTS idx_minion_logs_minion ON minion_logs(minion_id);
    CREATE INDEX IF NOT EXISTS idx_minion_logs_project ON minion_logs(project_id);

    CREATE TABLE IF NOT EXISTS session_meta (
      project_id TEXT PRIMARY KEY,
      claude_session_id TEXT,
      updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
  `)
}

// --- Messages ---

export async function saveMessage(projectId: string, msg: {
  id: string; role: string; content: string; timestamp: string
}): Promise<void> {
  const p = await getPool()
  await p.query(
    `INSERT INTO messages (id, project_id, role, content, timestamp)
     VALUES ($1, $2, $3, $4, $5)
     ON CONFLICT (id) DO UPDATE SET content = $4, timestamp = $5`,
    [msg.id, projectId, msg.role, msg.content, msg.timestamp]
  )
}

export async function getMessages(projectId: string): Promise<Array<{
  id: string; role: string; content: string; timestamp: string
}>> {
  const p = await getPool()
  const { rows } = await p.query(
    'SELECT id, role, content, timestamp FROM messages WHERE project_id = $1 ORDER BY created_at ASC',
    [projectId]
  )
  return rows
}

export async function deleteMessages(projectId: string): Promise<void> {
  const p = await getPool()
  await p.query('DELETE FROM messages WHERE project_id = $1', [projectId])
}

// --- Minions ---

export async function saveMinion(projectId: string, minion: {
  id: string; name: string; task: string; branch: string;
  worktreePath: string; projectPath: string; status: string;
  claudeSessionId: string | null
}): Promise<void> {
  const p = await getPool()
  await p.query(
    `INSERT INTO minions (id, project_id, name, task, branch, worktree_path, project_path, status, claude_session_id)
     VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
     ON CONFLICT (id) DO UPDATE SET status = $8, claude_session_id = $9`,
    [minion.id, projectId, minion.name, minion.task, minion.branch,
      minion.worktreePath, minion.projectPath, minion.status, minion.claudeSessionId]
  )
}

export async function getMinions(projectId: string): Promise<Array<{
  id: string; name: string; task: string; branch: string;
  worktree_path: string; project_path: string; status: string;
  claude_session_id: string | null
}>> {
  const p = await getPool()
  const { rows } = await p.query(
    'SELECT id, name, task, branch, worktree_path, project_path, status, claude_session_id FROM minions WHERE project_id = $1 ORDER BY created_at ASC',
    [projectId]
  )
  return rows
}

export async function deleteMinionsForProject(projectId: string): Promise<void> {
  const p = await getPool()
  await p.query('DELETE FROM minion_logs WHERE project_id = $1', [projectId])
  await p.query('DELETE FROM minions WHERE project_id = $1', [projectId])
}

// --- Minion Logs ---

export async function saveMinionLog(minionId: string, projectId: string, log: {
  id: string; role: string; content: string; timestamp: string
}): Promise<void> {
  const p = await getPool()
  await p.query(
    `INSERT INTO minion_logs (id, minion_id, project_id, role, content, timestamp)
     VALUES ($1, $2, $3, $4, $5, $6)
     ON CONFLICT (id) DO UPDATE SET content = $5`,
    [log.id, minionId, projectId, log.role, log.content, log.timestamp]
  )
}

export async function getMinionLogs(minionId: string): Promise<Array<{
  id: string; role: string; content: string; timestamp: string
}>> {
  const p = await getPool()
  const { rows } = await p.query(
    'SELECT id, role, content, timestamp FROM minion_logs WHERE minion_id = $1 ORDER BY created_at ASC',
    [minionId]
  )
  return rows
}

export async function getMinionLogsForProject(minionId: string, projectId: string): Promise<Array<{
  id: string; role: string; content: string; timestamp: string
}>> {
  return getMinionLogs(minionId)
}

// --- Session Meta ---

export async function saveSessionMeta(projectId: string, claudeSessionId: string | null): Promise<void> {
  const p = await getPool()
  await p.query(
    `INSERT INTO session_meta (project_id, claude_session_id, updated_at)
     VALUES ($1, $2, NOW())
     ON CONFLICT (project_id) DO UPDATE SET claude_session_id = $2, updated_at = NOW()`,
    [projectId, claudeSessionId]
  )
}

export async function getSessionMeta(projectId: string): Promise<{ claude_session_id: string | null } | null> {
  const p = await getPool()
  const { rows } = await p.query(
    'SELECT claude_session_id FROM session_meta WHERE project_id = $1',
    [projectId]
  )
  return rows[0] ?? null
}

// --- Full session clear ---

export async function clearProjectSession(projectId: string): Promise<void> {
  const p = await getPool()
  await p.query('DELETE FROM minion_logs WHERE project_id = $1', [projectId])
  await p.query('DELETE FROM minions WHERE project_id = $1', [projectId])
  await p.query('DELETE FROM messages WHERE project_id = $1', [projectId])
  await p.query('DELETE FROM session_meta WHERE project_id = $1', [projectId])
}
