import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createOAuthState } from "@/lib/oauth-state"

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
  const label = rawLabel.trim().slice(0, 100) || "GitHub Account"

  const state = createOAuthState(user.id, label)
  const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000"
  const redirectUri = `${siteUrl}/api/integrations/github/callback`

  const params = new URLSearchParams({
    client_id: process.env.GITHUB_INTEGRATION_CLIENT_ID!,
    redirect_uri: redirectUri,
    scope: "repo gist read:org workflow user:email",
    state,
  })

  return NextResponse.redirect(
    `https://github.com/login/oauth/authorize?${params.toString()}`
  )
}
