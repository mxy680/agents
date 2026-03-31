import {
  saveMessage,
  getMessages,
  saveMinion,
  getMinions,
  saveMinionLog,
  getMinionLogs,
  saveSessionMeta,
  getSessionMeta,
  clearProjectSession
} from './database'

interface MessageData {
  id: string
  role: string
  content: string
  timestamp: string
}

interface MinionLogData {
  id: string
  role: string
  content: string
  timestamp: string
}

interface MinionData {
  id: string
  name: string
  task: string
  branch: string
  worktreePath: string
  projectPath: string
  status: string
  claudeSessionId: string | null
  logs: MinionLogData[]
}

interface SessionData {
  messages: MessageData[]
  minions: MinionData[]
  claudeSessionId: string | null
}

export async function saveSession(
  _event: unknown,
  input: { projectId: string; data: SessionData }
): Promise<{ error: string | null }> {
  try {
    const { projectId, data } = input

    await saveSessionMeta(projectId, data.claudeSessionId)

    for (const msg of data.messages) {
      await saveMessage(projectId, msg)
    }

    for (const minion of data.minions) {
      await saveMinion(projectId, minion)
      for (const log of minion.logs) {
        await saveMinionLog(minion.id, projectId, log)
      }
    }

    return { error: null }
  } catch (err) {
    return { error: err instanceof Error ? err.message : 'Failed to save session' }
  }
}

export async function loadSession(
  _event: unknown,
  input: { projectId: string }
): Promise<{ data: SessionData | null; error: string | null }> {
  try {
    const { projectId } = input

    const meta = await getSessionMeta(projectId)
    const messages = await getMessages(projectId)
    const minionRows = await getMinions(projectId)

    if (messages.length === 0 && minionRows.length === 0 && !meta) {
      return { data: null, error: null }
    }

    const minions: MinionData[] = []
    for (const m of minionRows) {
      const logs = await getMinionLogs(m.id)
      minions.push({
        id: m.id,
        name: m.name,
        task: m.task,
        branch: m.branch,
        worktreePath: m.worktree_path,
        projectPath: m.project_path,
        status: m.status,
        claudeSessionId: m.claude_session_id,
        logs
      })
    }

    return {
      data: {
        messages,
        minions,
        claudeSessionId: meta?.claude_session_id ?? null
      },
      error: null
    }
  } catch (err) {
    return { data: null, error: err instanceof Error ? err.message : 'Failed to load session' }
  }
}

export async function deleteSession(
  _event: unknown,
  input: { projectId: string }
): Promise<{ error: string | null }> {
  try {
    await clearProjectSession(input.projectId)
    return { error: null }
  } catch (err) {
    return { error: err instanceof Error ? err.message : 'Failed to delete session' }
  }
}
