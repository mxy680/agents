import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { encrypt } from "@/lib/crypto"
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

  const body = await request.json().catch(() => ({}))
  const serverUrl = typeof body.url === "string" ? body.url.trim() : ""
  const password = typeof body.password === "string" ? body.password.trim() : ""
  const label = typeof body.label === "string" ? body.label.trim().slice(0, 100) : "iMessage (BlueBubbles)"

  if (!serverUrl) {
    return NextResponse.json({ error: "Server URL is required" }, { status: 400 })
  }
  if (!password) {
    return NextResponse.json({ error: "Server password is required" }, { status: 400 })
  }

  // Validate URL format
  try {
    new URL(serverUrl)
  } catch {
    return NextResponse.json({ error: "Invalid server URL" }, { status: 400 })
  }

  // Verify connectivity by pinging the BlueBubbles server
  try {
    const pingUrl = new URL("/api/v1/ping", serverUrl)
    pingUrl.searchParams.set("password", password)
    const pingRes = await fetch(pingUrl.toString(), {
      method: "GET",
      signal: AbortSignal.timeout(10000),
    })
    if (!pingRes.ok) {
      return NextResponse.json(
        { error: `BlueBubbles server returned ${pingRes.status}. Check your URL and password.` },
        { status: 400 }
      )
    }
  } catch {
    return NextResponse.json(
      { error: "Could not reach BlueBubbles server. Ensure it is running and accessible." },
      { status: 400 }
    )
  }

  const credentials = JSON.stringify({ url: serverUrl, password })
  const encrypted = encrypt(credentials)

  const admin = createAdminClient()
  const { error: dbError } = await admin
    .from("user_integrations")
    .upsert(
      {
        user_id: user.id,
        provider: "bluebubbles",
        label,
        status: "active",
        credentials: `\\x${Buffer.from(encrypted).toString("hex")}`,
        updated_at: new Date().toISOString(),
      },
      { onConflict: "user_id,provider,label" }
    )

  if (dbError) {
    console.error("[bluebubbles/save] DB error:", dbError)
    return NextResponse.json({ error: "Failed to save credentials" }, { status: 500 })
  }

  return NextResponse.json({ success: true })
}
