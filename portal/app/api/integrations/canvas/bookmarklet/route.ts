import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { encrypt } from "@/lib/crypto"

/**
 * POST /api/integrations/canvas/bookmarklet
 *
 * Receives Canvas cookies from the bookmarklet popup window.
 * Auth: user must be logged into the portal (session cookie).
 *
 * Body: { base_url, cookies: { cookie_name: value, ... }, label? }
 */
export async function POST(request: NextRequest) {
  const supabase = await createClient()
  const {
    data: { user },
  } = await supabase.auth.getUser()

  if (!user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  const body = await request.json().catch(() => null)
  if (!body || typeof body.base_url !== "string" || typeof body.cookies !== "object") {
    return NextResponse.json({ error: "Invalid request body" }, { status: 400 })
  }

  const baseUrl = body.base_url.trim().replace(/\/+$/, "")
  const rawCookies = body.cookies as Record<string, string>
  const label = typeof body.label === "string" ? body.label.trim().slice(0, 100) : "Canvas LMS"

  if (!baseUrl) {
    return NextResponse.json({ error: "Canvas URL is required" }, { status: 400 })
  }

  // Find session cookie (try known names)
  const sessionCookie =
    rawCookies["_normandy_session"] ||
    rawCookies["canvas_session"] ||
    rawCookies["_legacy_normandy_session"] ||
    null

  if (!sessionCookie) {
    return NextResponse.json(
      { error: "No Canvas session cookie found. Make sure you are logged in." },
      { status: 400 }
    )
  }

  // Build credential payload
  const credentials: Record<string, string> = {
    base_url: baseUrl,
    session_cookie: sessionCookie,
  }

  if (rawCookies["_csrf_token"]) {
    credentials.csrf_token = rawCookies["_csrf_token"]
  }
  if (rawCookies["log_session_id"]) {
    credentials.log_session_id = rawCookies["log_session_id"]
  }

  // Verify the session works by hitting Canvas API
  try {
    const cookieHeader = Object.entries(rawCookies)
      .map(([k, v]) => `${k}=${v}`)
      .join("; ")

    const verifyRes = await fetch(`${baseUrl}/api/v1/users/self`, {
      headers: {
        Cookie: cookieHeader,
        Accept: "application/json",
      },
      signal: AbortSignal.timeout(10000),
    })

    if (!verifyRes.ok) {
      return NextResponse.json(
        { error: `Canvas returned ${verifyRes.status}. Session may have expired.` },
        { status: 400 }
      )
    }
  } catch {
    return NextResponse.json(
      { error: "Could not reach Canvas. Check your URL and try again." },
      { status: 400 }
    )
  }

  // Encrypt and save
  const encrypted = encrypt(JSON.stringify(credentials))

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
    console.error("[canvas/bookmarklet] DB error:", dbError)
    return NextResponse.json({ error: "Failed to save credentials" }, { status: 500 })
  }

  return NextResponse.json({ success: true })
}
