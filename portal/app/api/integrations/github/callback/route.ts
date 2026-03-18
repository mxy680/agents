import { NextRequest, NextResponse } from "next/server"
import { encrypt } from "@/lib/crypto"
import { createAdminClient } from "@/lib/supabase/admin"

interface GitHubTokenResponse {
  access_token: string
  token_type: string
  scope: string
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

  const colonIdx = state.indexOf(":")
  if (colonIdx === -1) {
    return NextResponse.redirect(`${siteUrl}/integrations?error=invalid_state`)
  }

  const userId = state.slice(0, colonIdx)
  const label = state.slice(colonIdx + 1)

  const siteUrlForRedirect = process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000"
  const redirectUri = `${siteUrlForRedirect}/api/integrations/github/callback`

  const tokenRes = await fetch("https://github.com/login/oauth/access_token", {
    method: "POST",
    headers: {
      "Content-Type": "application/x-www-form-urlencoded",
      Accept: "application/json",
    },
    body: new URLSearchParams({
      client_id: process.env.GITHUB_INTEGRATION_CLIENT_ID!,
      client_secret: process.env.GITHUB_INTEGRATION_CLIENT_SECRET!,
      code,
      redirect_uri: redirectUri,
    }),
  })

  if (!tokenRes.ok) {
    console.error("GitHub token exchange failed:", await tokenRes.text())
    return NextResponse.redirect(`${siteUrl}/integrations?error=token_exchange_failed`)
  }

  const tokens: GitHubTokenResponse = await tokenRes.json()

  if (tokens.error || !tokens.access_token) {
    console.error("GitHub token error:", tokens.error_description)
    return NextResponse.redirect(`${siteUrl}/integrations?error=token_exchange_failed`)
  }

  const payload = JSON.stringify({
    access_token: tokens.access_token,
  })

  const encrypted = encrypt(payload)

  const admin = createAdminClient()
  const { error: dbError } = await admin.from("user_integrations").insert({
    user_id: userId,
    provider: "github",
    label,
    status: "active",
    credentials: encrypted,
  })

  if (dbError) {
    console.error("Failed to save GitHub integration:", dbError)
    return NextResponse.redirect(`${siteUrl}/integrations?error=save_failed`)
  }

  return NextResponse.redirect(`${siteUrl}/integrations`)
}
