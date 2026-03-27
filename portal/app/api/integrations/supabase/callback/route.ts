import { NextRequest, NextResponse } from "next/server"
import { encrypt } from "@/lib/crypto"
import { createAdminClient } from "@/lib/supabase/admin"
import { createClient } from "@/lib/supabase/server"
import { verifyOAuthState } from "@/lib/oauth-state"

interface SupabaseTokenResponse {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
  error?: string
  error_description?: string
}

export async function GET(request: NextRequest) {
  const { searchParams } = new URL(request.url)
  const code = searchParams.get("code")
  const state = searchParams.get("state")
  const error = searchParams.get("error")

  const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000"

  if (error || !code || !state) {
    return NextResponse.redirect(`${siteUrl}/integrations?error=oauth_denied`)
  }

  let userId: string
  let label: string
  try {
    const verified = verifyOAuthState(state)
    userId = verified.userId
    label = verified.label
  } catch {
    return NextResponse.redirect(`${siteUrl}/integrations?error=invalid_state`)
  }

  // Use the mock admin user ID (auth is local-only, no session needed)
  userId = "00000000-0000-0000-0000-000000000001"

  // Retrieve PKCE code_verifier from cookie
  const codeVerifier = request.cookies.get("supabase_pkce_verifier")?.value
  if (!codeVerifier) {
    return NextResponse.redirect(`${siteUrl}/integrations?error=missing_pkce_verifier`)
  }

  const redirectUri = `${siteUrl}/api/integrations/supabase/callback`

  const tokenRes = await fetch("https://api.supabase.com/v1/oauth/token", {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
      Accept: "application/json",
    },
    body: new URLSearchParams({
      grant_type: "authorization_code",
      client_id: process.env.SUPABASE_INTEGRATION_CLIENT_ID!,
      client_secret: process.env.SUPABASE_INTEGRATION_CLIENT_SECRET!,
      code,
      redirect_uri: redirectUri,
      code_verifier: codeVerifier,
    }),
  })

  if (!tokenRes.ok) {
    const errText = await tokenRes.text()
    console.error("[supabase-cb] Token exchange failed:", tokenRes.status, errText)
    return NextResponse.redirect(`${siteUrl}/integrations?error=token_exchange_failed`)
  }

  const tokens: SupabaseTokenResponse = await tokenRes.json()
  console.log("[supabase-cb] Token exchange OK, has access_token:", !!tokens.access_token)

  if (tokens.error || !tokens.access_token) {
    console.error("[supabase-cb] Token error:", tokens.error_description)
    return NextResponse.redirect(`${siteUrl}/integrations?error=token_exchange_failed`)
  }

  const payload = JSON.stringify({
    access_token: tokens.access_token,
    refresh_token: tokens.refresh_token,
  })

  console.log("[supabase-cb] Encrypting credentials, userId:", userId, "label:", label)
  const encrypted = encrypt(payload)

  const admin = createAdminClient()
  console.log("[supabase-cb] Upserting to user_integrations...")
  const { error: dbError } = await admin.from("user_integrations").upsert(
    {
      user_id: userId,
      provider: "supabase",
      label,
      status: "active",
      credentials: `\\x${encrypted.toString("hex")}`,
      updated_at: new Date().toISOString(),
    },
    { onConflict: "user_id,provider,label" }
  )

  if (dbError) {
    console.error("[supabase-cb] DB SAVE FAILED:", JSON.stringify(dbError))
    return NextResponse.redirect(`${siteUrl}/integrations?error=save_failed`)
  }
  console.log("[supabase-cb] Save succeeded!")

  // Clear the PKCE cookie
  const response = NextResponse.redirect(`${siteUrl}/integrations`)
  response.cookies.delete("supabase_pkce_verifier")
  return response
}
