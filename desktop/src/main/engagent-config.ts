import { join } from 'path'
import { execFileSync } from 'child_process'
import { homedir } from 'os'

let cachedEnv: Record<string, string> | null = null
let cacheTime = 0
const CACHE_TTL = 5 * 60 * 1000 // 5 minutes

function fetchDopplerEnv(): Record<string, string> {
  const now = Date.now()
  if (cachedEnv && now - cacheTime < CACHE_TTL) return cachedEnv

  try {
    const stdout = execFileSync('doppler', [
      'secrets', 'download',
      '--project', 'agents',
      '--config', 'dev',
      '--format', 'json',
      '--no-file'
    ], { timeout: 10000, encoding: 'utf-8' })

    cachedEnv = JSON.parse(stdout)
    cacheTime = now
    return cachedEnv!
  } catch {
    // Doppler not available — return cached or empty
    return cachedEnv ?? {}
  }
}

/**
 * Returns environment variables to inject into Claude Code processes
 * so the engagent CLI can resolve credentials from the database.
 * Pulls from Doppler (agents/dev config) as the single source of truth.
 */
export function getEngagentEnv(): Record<string, string> {
  try {
    const doppler = fetchDopplerEnv()
    const env: Record<string, string> = {}

    // Credential resolution
    if (doppler.ENCRYPTION_MASTER_KEY) {
      env.RESOLVE_CREDENTIALS = '1'
      env.SUPABASE_DB_URL = doppler.SUPABASE_DB_URL ?? ''
      env.ENCRYPTION_MASTER_KEY = doppler.ENCRYPTION_MASTER_KEY
    }

    // Google OAuth (for token refresh)
    if (doppler.GOOGLE_CLIENT_ID) {
      env.GOOGLE_CLIENT_ID = doppler.GOOGLE_CLIENT_ID
      env.GOOGLE_CLIENT_SECRET = doppler.GOOGLE_CLIENT_SECRET ?? ''
    }

    // GitHub OAuth (for token refresh)
    if (doppler.GITHUB_INTEGRATION_CLIENT_ID) {
      env.GITHUB_CLIENT_ID = doppler.GITHUB_INTEGRATION_CLIENT_ID
      env.GITHUB_CLIENT_SECRET = doppler.GITHUB_INTEGRATION_CLIENT_SECRET ?? ''
    }

    // Supabase OAuth (for token refresh)
    if (doppler.SUPABASE_INTEGRATION_CLIENT_ID) {
      env.SUPABASE_INTEGRATION_CLIENT_ID = doppler.SUPABASE_INTEGRATION_CLIENT_ID
      env.SUPABASE_INTEGRATION_CLIENT_SECRET = doppler.SUPABASE_INTEGRATION_CLIENT_SECRET ?? ''
    }

    // GCP
    if (doppler.GCP_PROJECT_ID) {
      env.GCP_PROJECT_ID = doppler.GCP_PROJECT_ID
    }
    if (doppler.GCP_SERVICE_ACCOUNT_JSON) {
      env.GCP_SERVICE_ACCOUNT_JSON = doppler.GCP_SERVICE_ACCOUNT_JSON
    }

    // Add CLI binary to PATH
    const binDir = join(homedir(), '.ade', 'bin')
    const currentPath = process.env.PATH || ''
    env.PATH = `${binDir}:${currentPath}`

    return env
  } catch {
    return {}
  }
}

// IPC handlers (simplified — just expose what Doppler has)
export function getEngagentConfigHandler(): Record<string, string> {
  const doppler = fetchDopplerEnv()
  // Return safe subset — no raw secrets
  return {
    hasDoppler: Object.keys(doppler).length > 0 ? 'true' : 'false',
    project: doppler.DOPPLER_PROJECT ?? '',
    config: doppler.DOPPLER_CONFIG ?? '',
    hasEncryptionKey: doppler.ENCRYPTION_MASTER_KEY ? 'true' : 'false',
    hasGoogleOAuth: doppler.GOOGLE_CLIENT_ID ? 'true' : 'false',
    hasGitHubOAuth: doppler.GITHUB_INTEGRATION_CLIENT_ID ? 'true' : 'false',
    hasSupabaseOAuth: doppler.SUPABASE_INTEGRATION_CLIENT_ID ? 'true' : 'false',
    databaseUrl: doppler.SUPABASE_DB_URL ? 'configured' : 'missing'
  }
}

export function saveEngagentConfigHandler(
  _event: unknown,
  _config: Record<string, unknown>
): { error: string | null } {
  return { error: 'Config is managed by Doppler. Edit via `doppler secrets set` or the Doppler dashboard.' }
}
