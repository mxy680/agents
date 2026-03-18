import { NextRequest, NextResponse } from "next/server"
import { encrypt } from "@/lib/crypto"
import { createAdminClient } from "@/lib/supabase/admin"
import { createClient } from "@/lib/supabase/server"
import { verifyOAuthState } from "@/lib/oauth-state"

interface GoogleTokenResponse {
  access_token: string
  refresh_token?: string
  expires_in: number
  token_type: string
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

  // Verify the authenticated session user matches the state userId
  try {
    const supabase = await createClient()
    const {
      data: { user },
    } = await supabase.auth.getUser()
    if (!user || user.id !== userId) {
      return NextResponse.redirect(`${siteUrl}/integrations?error=session_mismatch`)
    }
  } catch {
    return NextResponse.redirect(`${siteUrl}/integrations?error=session_check_failed`)
  }

  const redirectUri = `${siteUrl}/api/integrations/google/callback`

  const tokenRes = await fetch("https://oauth2.googleapis.com/token", {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      code,
      client_id: process.env.GOOGLE_CLIENT_ID!,
      client_secret: process.env.GOOGLE_CLIENT_SECRET!,
      redirect_uri: redirectUri,
      grant_type: "authorization_code",
    }),
  })

  if (!tokenRes.ok) {
    console.error("Google token exchange failed:", await tokenRes.text())
    return NextResponse.redirect(`${siteUrl}/integrations?error=token_exchange_failed`)
  }

  const tokens: GoogleTokenResponse = await tokenRes.json()

  const payload = JSON.stringify({
    access_token: tokens.access_token,
    refresh_token: tokens.refresh_token ?? "",
    token_expiry: new Date(Date.now() + tokens.expires_in * 1000).toISOString(),
  })

  const encrypted = encrypt(payload)

  const admin = createAdminClient()
  const { error: dbError } = await admin.from("user_integrations").upsert(
    {
      user_id: userId,
      provider: "google",
      label,
      status: "active",
      credentials: `\\x${encrypted.toString("hex")}`,
      updated_at: new Date().toISOString(),
    },
    { onConflict: "user_id,provider,label" }
  )

  if (dbError) {
    console.error("Failed to save Google integration:", dbError)
    return NextResponse.redirect(`${siteUrl}/integrations?error=save_failed`)
  }

  return NextResponse.redirect(`${siteUrl}/integrations`)
}
