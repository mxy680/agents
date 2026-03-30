import { createAdminClient } from "@/lib/supabase/admin"
import { decrypt } from "@/lib/crypto"

/**
 * Maps decrypted credential JSON to environment variables for a given provider.
 */
export function credentialsToEnv(provider: string, credJson: Record<string, string>): Record<string, string> {
  const env: Record<string, string> = {}
  switch (provider) {
    case "google":
      if (credJson.access_token) env.GOOGLE_ACCESS_TOKEN = credJson.access_token
      if (credJson.refresh_token) env.GOOGLE_REFRESH_TOKEN = credJson.refresh_token
      if (process.env.GOOGLE_CLIENT_ID) env.GOOGLE_CLIENT_ID = process.env.GOOGLE_CLIENT_ID
      if (process.env.GOOGLE_CLIENT_SECRET) env.GOOGLE_CLIENT_SECRET = process.env.GOOGLE_CLIENT_SECRET
      break
    case "github":
      if (credJson.access_token) env.GITHUB_ACCESS_TOKEN = credJson.access_token
      if (credJson.refresh_token) env.GITHUB_REFRESH_TOKEN = credJson.refresh_token
      // Use integration client (the one that issued the token) for refresh
      if (process.env.GITHUB_INTEGRATION_CLIENT_ID) env.GITHUB_CLIENT_ID = process.env.GITHUB_INTEGRATION_CLIENT_ID
      if (process.env.GITHUB_INTEGRATION_CLIENT_SECRET) env.GITHUB_CLIENT_SECRET = process.env.GITHUB_INTEGRATION_CLIENT_SECRET
      break
    case "instagram":
      if (credJson.session_id) env.INSTAGRAM_SESSION_ID = credJson.session_id
      if (credJson.csrf_token) env.INSTAGRAM_CSRF_TOKEN = credJson.csrf_token
      if (credJson.ds_user_id) env.INSTAGRAM_DS_USER_ID = credJson.ds_user_id
      break
    case "linkedin":
      if (credJson.li_at) env.LINKEDIN_LI_AT = credJson.li_at
      if (credJson.jsessionid) env.LINKEDIN_JSESSIONID = credJson.jsessionid
      if (credJson.bcookie) env.LINKEDIN_BCOOKIE = credJson.bcookie
      if (credJson.lidc) env.LINKEDIN_LIDC = credJson.lidc
      if (credJson.li_mc) env.LINKEDIN_LI_MC = credJson.li_mc
      break
    case "framer":
      if (credJson.api_key) env.FRAMER_API_KEY = credJson.api_key
      if (credJson.project_url) env.FRAMER_PROJECT_URL = credJson.project_url
      break
    case "supabase":
      if (credJson.access_token) env.SUPABASE_ACCESS_TOKEN = credJson.access_token
      if (credJson.refresh_token) env.SUPABASE_REFRESH_TOKEN = credJson.refresh_token
      if (process.env.SUPABASE_INTEGRATION_CLIENT_ID) env.SUPABASE_INTEGRATION_CLIENT_ID = process.env.SUPABASE_INTEGRATION_CLIENT_ID
      if (process.env.SUPABASE_INTEGRATION_CLIENT_SECRET) env.SUPABASE_INTEGRATION_CLIENT_SECRET = process.env.SUPABASE_INTEGRATION_CLIENT_SECRET
      break
    case "x":
      if (credJson.auth_token) env.X_AUTH_TOKEN = credJson.auth_token
      if (credJson.csrf_token) env.X_CSRF_TOKEN = credJson.csrf_token
      break
    case "canvas":
      // Go CLI expects CANVAS_COOKIES (full cookie string) and CANVAS_BASE_URL
      if (credJson.all_cookies) env.CANVAS_COOKIES = credJson.all_cookies
      if (credJson.base_url) env.CANVAS_BASE_URL = credJson.base_url
      else if (process.env.CANVAS_BASE_URL) env.CANVAS_BASE_URL = process.env.CANVAS_BASE_URL
      break
    case "bluebubbles":
      if (credJson.url) env.BLUEBUBBLES_URL = credJson.url
      if (credJson.password) env.BLUEBUBBLES_PASSWORD = credJson.password
      break
    case "zillow":
      if (credJson.all_cookies) env.ZILLOW_COOKIES = credJson.all_cookies
      if (credJson.proxy_url) env.ZILLOW_PROXY_URL = credJson.proxy_url
      break
    case "vercel":
      if (credJson.token) env.VERCEL_TOKEN = credJson.token
      if (credJson.team_id) env.VERCEL_TEAM_ID = credJson.team_id
      break
    case "cloudflare":
      if (credJson.token) env.CLOUDFLARE_API_TOKEN = credJson.token
      if (credJson.account_id) env.CLOUDFLARE_ACCOUNT_ID = credJson.account_id
      break
    case "linear":
      if (credJson.token) env.LINEAR_API_KEY = credJson.token
      break
    case "fly":
      if (credJson.token) env.FLY_API_TOKEN = credJson.token
      break
    case "gcp":
      if (credJson.token) env.GCP_ACCESS_TOKEN = credJson.token
      if (credJson.service_account_json) env.GCP_SERVICE_ACCOUNT_JSON = credJson.service_account_json
      if (credJson.project_id) env.GCP_PROJECT_ID = credJson.project_id
      break
    case "gcp-console":
      if (credJson.all_cookies) env.GCP_CONSOLE_ALL_COOKIES = credJson.all_cookies
      if (credJson.SAPISID) env.GCP_CONSOLE_SAPISID = credJson.SAPISID
      break  }
  return env
}

/**
 * Resolves all active integration credentials for the admin user.
 * Single-tenant: all integrations are owned centrally.
 */
export async function resolveAdminCredentials(): Promise<Record<string, string>> {
  const adminUserId = process.env.ADMIN_USER_ID
  if (adminUserId) {
    return resolveUserCredentials(adminUserId)
  }
  // Fallback: resolve all active integrations regardless of owner
  const admin = createAdminClient()
  const { data: integrations, error } = await admin
    .from("user_integrations")
    .select("provider, credentials, status")
    .eq("status", "active")

  if (error) {
    throw new Error(`Failed to fetch integrations: ${error.message}`)
  }

  const credEnv: Record<string, string> = {}
  for (const integration of integrations ?? []) {
    try {
      const raw = integration.credentials
      let buf: Buffer
      if (typeof raw === "string") {
        const hex = raw.startsWith("\\x") ? raw.slice(2) : raw
        buf = Buffer.from(hex, "hex")
      } else if (Buffer.isBuffer(raw)) {
        buf = raw
      } else {
        buf = Buffer.from(raw as ArrayBuffer)
      }
      const decrypted = decrypt(buf)
      const credJson = JSON.parse(decrypted) as Record<string, string>
      const envVars = credentialsToEnv(integration.provider, credJson)
      Object.assign(credEnv, envVars)
    } catch {
      console.error(`[credentials] Failed to decrypt credentials for ${integration.provider}`)
    }
  }

  return credEnv
}

/**
 * Resolves all active integration credentials for a user into environment variables.
 * @deprecated Use resolveAdminCredentials() for single-tenant operation.
 */
export async function resolveUserCredentials(userId: string): Promise<Record<string, string>> {
  const admin = createAdminClient()
  const { data: integrations, error } = await admin
    .from("user_integrations")
    .select("provider, credentials, status")
    .eq("user_id", userId)
    .eq("status", "active")

  if (error) {
    throw new Error(`Failed to fetch integrations: ${error.message}`)
  }

  const credEnv: Record<string, string> = {}
  for (const integration of integrations ?? []) {
    try {
      const raw = integration.credentials
      let buf: Buffer
      if (typeof raw === "string") {
        const hex = raw.startsWith("\\x") ? raw.slice(2) : raw
        buf = Buffer.from(hex, "hex")
      } else if (Buffer.isBuffer(raw)) {
        buf = raw
      } else {
        buf = Buffer.from(raw as ArrayBuffer)
      }
      const decrypted = decrypt(buf)
      const credJson = JSON.parse(decrypted) as Record<string, string>
      const envVars = credentialsToEnv(integration.provider, credJson)
      Object.assign(credEnv, envVars)
    } catch {
      console.error(`[credentials] Failed to decrypt credentials for ${integration.provider}`)
    }
  }

  return credEnv
}
