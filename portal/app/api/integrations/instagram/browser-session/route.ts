import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createOAuthState } from "@/lib/oauth-state"

export async function POST(request: NextRequest) {
  const supabase = await createClient()
  const {
    data: { user },
  } = await supabase.auth.getUser()

  if (!user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  const body = await request.json().catch(() => ({}))
  const label =
    typeof body.label === "string" && body.label.trim()
      ? body.label.trim()
      : "Instagram Account"

  // Create a signed token embedding userId and label.
  // The WS server verifies this token on connection.
  const token = createOAuthState(user.id, label)

  const wsPort = process.env.BROWSER_WS_PORT ?? "3001"
  const wsHost = process.env.BROWSER_WS_HOST ?? "localhost"
  const wsUrl = `ws://${wsHost}:${wsPort}/browser-session?token=${encodeURIComponent(token)}`

  return NextResponse.json({ wsUrl })
}
