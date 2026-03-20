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
  }
  return env
}

/**
 * Resolves all active integration credentials for a user into environment variables.
 * Used by both the chat API and job runner.
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
      console.error(`[credentials] ${integration.provider}: decrypted OK, env keys: ${Object.keys(envVars).join(", ")}`)
    } catch (e) {
      console.error(`[credentials] Failed to decrypt credentials for ${integration.provider}:`, e)
    }
  }

  console.error(`[credentials] Total env vars: ${Object.keys(credEnv).length} — keys: ${Object.keys(credEnv).join(", ")}`)
  return credEnv
}
