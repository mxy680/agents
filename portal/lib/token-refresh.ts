import { createAdminClient } from "@/lib/supabase/admin"
import { encrypt, decrypt } from "@/lib/crypto"

/** Providers that support OAuth token refresh. */
const REFRESH_PROVIDERS = new Set(["google", "supabase"])

/** Refresh if token expires within this window. */
const REFRESH_BUFFER_MS = 10 * 60 * 1000 // 10 minutes

interface RefreshResult {
  access_token: string
  refresh_token: string
  token_expiry: string
}

/**
 * Returns true if the credential's access token should be refreshed.
 * Triggers refresh when token_expiry is missing or within REFRESH_BUFFER_MS of now.
 */
export function needsRefresh(provider: string, credJson: Record<string, string>): boolean {
  if (!REFRESH_PROVIDERS.has(provider)) return false
  if (!credJson.refresh_token) return false

  const expiry = credJson.token_expiry
  if (!expiry) return true // no expiry recorded → assume expired

  const expiresAt = new Date(expiry).getTime()
  return Date.now() + REFRESH_BUFFER_MS >= expiresAt
}

/**
 * Refreshes an OAuth access token for the given provider.
 * Returns null if the provider doesn't support refresh or refresh fails.
 */
export async function refreshOAuthToken(
  provider: string,
  credJson: Record<string, string>
): Promise<RefreshResult | null> {
  if (!REFRESH_PROVIDERS.has(provider)) return null
  if (!credJson.refresh_token) return null

  switch (provider) {
    case "google":
      return refreshGoogle(credJson.refresh_token)
    case "supabase":
      return refreshSupabase(credJson.refresh_token)
    default:
      return null
  }
}

async function refreshGoogle(refreshToken: string): Promise<RefreshResult> {
  const clientId = process.env.GOOGLE_CLIENT_ID
  const clientSecret = process.env.GOOGLE_CLIENT_SECRET
  if (!clientId || !clientSecret) {
    throw new Error("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET required for refresh")
  }

  const res = await fetch("https://oauth2.googleapis.com/token", {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      client_id: clientId,
      client_secret: clientSecret,
      refresh_token: refreshToken,
      grant_type: "refresh_token",
    }),
  })

  if (!res.ok) {
    const text = await res.text()
    throw new Error(`Google token refresh failed (${res.status}): ${text}`)
  }

  const data = await res.json() as {
    access_token: string
    expires_in: number
    refresh_token?: string
  }

  return {
    access_token: data.access_token,
    // Google may rotate the refresh token
    refresh_token: data.refresh_token ?? refreshToken,
    token_expiry: new Date(Date.now() + data.expires_in * 1000).toISOString(),
  }
}

async function refreshSupabase(refreshToken: string): Promise<RefreshResult> {
  const clientId = process.env.SUPABASE_INTEGRATION_CLIENT_ID
  const clientSecret = process.env.SUPABASE_INTEGRATION_CLIENT_SECRET
  if (!clientId || !clientSecret) {
    throw new Error("SUPABASE_INTEGRATION_CLIENT_ID and SUPABASE_INTEGRATION_CLIENT_SECRET required for refresh")
  }

  const res = await fetch("https://api.supabase.com/v1/oauth/token", {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
      Accept: "application/json",
    },
    body: new URLSearchParams({
      grant_type: "refresh_token",
      client_id: clientId,
      client_secret: clientSecret,
      refresh_token: refreshToken,
    }),
  })

  if (!res.ok) {
    const text = await res.text()
    throw new Error(`Supabase token refresh failed (${res.status}): ${text}`)
  }

  const data = await res.json() as {
    access_token: string
    refresh_token: string
    expires_in: number
  }

  return {
    access_token: data.access_token,
    refresh_token: data.refresh_token,
    token_expiry: new Date(Date.now() + data.expires_in * 1000).toISOString(),
  }
}

/**
 * Persists refreshed tokens back to the user_integrations row.
 * Decrypts existing credentials, merges in the new tokens, re-encrypts, and updates.
 */
export async function persistRefreshedCredentials(
  integrationId: string,
  existingCredJson: Record<string, string>,
  refreshed: RefreshResult
): Promise<void> {
  const merged: Record<string, string> = {
    ...existingCredJson,
    access_token: refreshed.access_token,
    refresh_token: refreshed.refresh_token,
    token_expiry: refreshed.token_expiry,
  }

  const encrypted = encrypt(JSON.stringify(merged))

  const admin = createAdminClient()
  const { error } = await admin
    .from("user_integrations")
    .update({
      credentials: `\\x${encrypted.toString("hex")}`,
      updated_at: new Date().toISOString(),
    })
    .eq("id", integrationId)

  if (error) {
    console.error(`[token-refresh] Failed to persist refreshed credentials for ${integrationId}:`, error.message)
  }
}
