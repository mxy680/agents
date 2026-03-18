import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createOAuthState } from "@/lib/oauth-state"
import { checkRateLimit } from "@/lib/rate-limit"

export async function POST(request: NextRequest) {
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
  const label = rawLabel.slice(0, 100) || "Instagram Account"

  // Create a signed token embedding userId and label.
  // The WS server verifies this token sent as the first message (not in URL).
  const token = createOAuthState(user.id, label)

  const wsPort = process.env.BROWSER_WS_PORT ?? "3001"
  const wsHost = process.env.BROWSER_WS_HOST ?? "localhost"

  // Use wss:// for non-local hosts
  const isLocal = wsHost === "localhost" || wsHost === "127.0.0.1"
  const wsScheme = isLocal ? "ws" : "wss"
  const wsUrl = `${wsScheme}://${wsHost}:${wsPort}/browser-session`

  return NextResponse.json({ wsUrl, token })
}
