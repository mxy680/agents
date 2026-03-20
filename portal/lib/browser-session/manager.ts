import { BrowserSession } from "./session"

const sessions = new Map<string, BrowserSession>()
const userSessions = new Map<string, string>() // userId -> sessionId

export function createSession(userId: string, label: string, provider: string): BrowserSession {
  // Destroy any existing session for this user
  const existingId = userSessions.get(userId)
  if (existingId) {
    destroySession(existingId)
  }

  const maxSessions = parseInt(process.env.MAX_BROWSER_SESSIONS ?? "10", 10)
  if (sessions.size >= maxSessions) {
    throw new Error("Maximum browser sessions reached")
  }

  const id = crypto.randomUUID()
  const session = new BrowserSession(id, userId, label, provider)
  sessions.set(id, session)
  userSessions.set(userId, id)
  return session
}

export function getSession(sessionId: string): BrowserSession | undefined {
  return sessions.get(sessionId)
}

export function destroySession(sessionId: string): void {
  const session = sessions.get(sessionId)
  if (session) {
    userSessions.delete(session.userId)
    sessions.delete(sessionId)
    session.destroy()
  }
}

export function destroyAllSessions(): void {
  for (const [id] of sessions.entries()) {
    destroySession(id)
  }
}

// Cleanup stale sessions every 30 seconds
setInterval(() => {
  for (const [id, session] of sessions.entries()) {
    if (session.destroyed) {
      userSessions.delete(session.userId)
      sessions.delete(id)
    }
  }
}, 30_000)
