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
  const baseUrl = typeof body.base_url === "string" ? body.base_url.trim() : ""
  const sessionCookie = typeof body.session_cookie === "string" ? body.session_cookie.trim() : ""
  const csrfToken = typeof body.csrf_token === "string" ? body.csrf_token.trim() : ""
  const logSessionId = typeof body.log_session_id === "string" ? body.log_session_id.trim() : ""
  const label = typeof body.label === "string" ? body.label.trim().slice(0, 100) : "Canvas LMS"

  if (!baseUrl) {
    return NextResponse.json({ error: "Canvas instance URL is required" }, { status: 400 })
  }
  if (!sessionCookie) {
    return NextResponse.json({ error: "Session cookie is required" }, { status: 400 })
  }
  if (!csrfToken) {
    return NextResponse.json({ error: "CSRF token is required" }, { status: 400 })
  }

  // Validate URL format
  try {
    new URL(baseUrl)
  } catch {
    return NextResponse.json({ error: "Invalid Canvas instance URL" }, { status: 400 })
  }

  // Verify connectivity by calling the Canvas API with the provided session cookies
  try {
    const verifyUrl = new URL("/api/v1/users/self", baseUrl)
    const verifyRes = await fetch(verifyUrl.toString(), {
      method: "GET",
      headers: {
        Cookie: `_normandy_session=${sessionCookie}; _csrf_token=${csrfToken}`,
        Accept: "application/json",
      },
      signal: AbortSignal.timeout(10000),
    })
    if (!verifyRes.ok) {
      return NextResponse.json(
        { error: `Canvas returned ${verifyRes.status}. Check your session cookies and Canvas URL.` },
        { status: 400 }
      )
    }
  } catch {
    return NextResponse.json(
      { error: "Could not reach Canvas. Ensure the URL is correct and the server is accessible." },
      { status: 400 }
    )
  }

  const credentialData: Record<string, string> = {
    base_url: baseUrl,
    session_cookie: sessionCookie,
    csrf_token: csrfToken,
  }
  if (logSessionId) {
    credentialData.log_session_id = logSessionId
  }

  const credentials = JSON.stringify(credentialData)
  const encrypted = encrypt(credentials)

  const admin = createAdminClient()
  const { error: dbError } = await admin
    .from("user_integrations")
    .upsert(
      {
        user_id: user.id,
        provider: "canvas",
        label,
        status: "active",
        credentials: `\\x${Buffer.from(encrypted).toString("hex")}`,
        updated_at: new Date().toISOString(),
      },
      { onConflict: "user_id,provider,label" }
    )

  if (dbError) {
    console.error("[canvas/save] DB error:", dbError)
    return NextResponse.json({ error: "Failed to save credentials" }, { status: 500 })
  }

  return NextResponse.json({ success: true })
}
