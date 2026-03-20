import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createOAuthState } from "@/lib/oauth-state"
import { checkRateLimit } from "@/lib/rate-limit"
import { checkOrigin } from "@/lib/csrf"

export async function POST(request: NextRequest) {
  const csrfError = checkOrigin(request)
  if (csrfError) return csrfError

  const supabase = await createClient()
  const {
    data: { user },
  } = await supabase.auth.getUser()

  if (!user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  // Rate limit: 3 sessions per minute per user
  if (!checkRateLimit(`browser-session:${user.id}`, 3, 60_000)) {
    return NextResponse.json({ error: "Too many requests" }, { status: 429 })
  }

  const body = await request.json().catch(() => ({}))
  const rawLabel = typeof body.label === "string" ? body.label.trim() : ""
  const label = rawLabel.slice(0, 100) || "LinkedIn Account"

  const token = createOAuthState(user.id, label)

  const wsHost = process.env.BROWSER_WS_HOST ?? "localhost"
  const isLocal = wsHost === "localhost" || wsHost === "127.0.0.1"

  const wsUrl = isLocal
    ? `ws://${wsHost}:${process.env.BROWSER_WS_PORT ?? "3001"}/browser-session/linkedin`
    : `wss://${wsHost}/ws/browser-session/linkedin`

  return NextResponse.json({ wsUrl, token })
}
