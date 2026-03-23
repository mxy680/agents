/**
 * Global session store that survives Next.js module reloading.
 * We attach it to globalThis so that both the connect and status
 * API routes share the same Map even with HMR / separate module instances.
 */

export type SessionStatus =
  | "pending"
  | "browser_open"
  | "waiting_login"
  | "capturing"
  | "done"
  | "saved"
  | "error";

export interface SessionProgress {
  sessionId: string;
  provider: string;
  status: SessionStatus;
  message: string;
  cookies?: Record<string, string>;
  error?: string;
}

const GLOBAL_KEY = "__emdash_playwright_sessions__" as const;

function getStore(): Map<string, SessionProgress> {
  const g = globalThis as Record<string, unknown>;
  if (!g[GLOBAL_KEY]) {
    g[GLOBAL_KEY] = new Map<string, SessionProgress>();
  }
  return g[GLOBAL_KEY] as Map<string, SessionProgress>;
}

export function getSession(sessionId: string): SessionProgress | undefined {
  return getStore().get(sessionId);
}

export function setSession(sessionId: string, progress: SessionProgress) {
  getStore().set(sessionId, progress);
}

export function updateSession(sessionId: string, update: Partial<SessionProgress>) {
  const current = getStore().get(sessionId);
  if (current) {
    getStore().set(sessionId, { ...current, ...update });
  }
}
