import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createOAuthState } from "@/lib/oauth-state"
import crypto from "crypto"

export async function GET(request: NextRequest) {
  const supabase = await createClient()
  const {
    data: { user },
  } = await supabase.auth.getUser()

  if (!user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  const { searchParams } = new URL(request.url)
  const rawLabel = searchParams.get("label") ?? ""
  const label = rawLabel.trim().slice(0, 100) || "Supabase Account"

  const state = createOAuthState(user.id, label)
  const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000"
  const redirectUri = `${siteUrl}/api/integrations/supabase/callback`

  // Supabase OAuth 2.1 requires PKCE (Proof Key for Code Exchange)
  const codeVerifier = crypto.randomBytes(32).toString("base64url")
  const codeChallenge = crypto
    .createHash("sha256")
    .update(codeVerifier)
    .digest("base64url")

  // Store code_verifier in a short-lived cookie for the callback
  const params = new URLSearchParams({
    client_id: process.env.SUPABASE_INTEGRATION_CLIENT_ID!,
    redirect_uri: redirectUri,
    response_type: "code",
    scope: "all",
    state,
    code_challenge: codeChallenge,
    code_challenge_method: "S256",
  })

  const response = NextResponse.redirect(
    `https://api.supabase.com/v1/oauth/authorize?${params.toString()}`
  )

  // Set code_verifier cookie (HttpOnly, Secure, short TTL)
  response.cookies.set("supabase_pkce_verifier", codeVerifier, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    maxAge: 600, // 10 minutes
    path: "/api/integrations/supabase",
  })

  return response
}
