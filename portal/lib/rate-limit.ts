const windows = new Map<string, { count: number; resetAt: number }>()

// Periodically clean up expired entries to prevent unbounded memory growth.
const CLEANUP_INTERVAL_MS = 60_000
let lastCleanup = Date.now()

function cleanupExpired(now: number) {
  if (now - lastCleanup < CLEANUP_INTERVAL_MS) return
  lastCleanup = now
  for (const [key, entry] of windows) {
    if (now > entry.resetAt) {
      windows.delete(key)
    }
  }
}

/**
 * Returns true if the request is within the allowed rate limit, false if it exceeds it.
 * Note: this is per-process — multiple serverless instances each have their own state.
 * @param key    Unique key per user/action (e.g. "browser-session:user-id")
 * @param max    Maximum requests allowed within the window
 * @param windowMs Window duration in milliseconds
 */
export function checkRateLimit(key: string, max: number, windowMs: number): boolean {
  const now = Date.now()
  cleanupExpired(now)
  const entry = windows.get(key)
  if (!entry || now > entry.resetAt) {
    windows.set(key, { count: 1, resetAt: now + windowMs })
    return true // allowed
  }
  entry.count++
  return entry.count <= max
}
