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
  const baseUrl = typeof body.base_url === "string" ? body.base_url.trim().replace(/\/+$/, "") : ""
  const rawCookies = typeof body.raw_cookies === "string" ? body.raw_cookies.trim() : ""
  const label = typeof body.label === "string" ? body.label.trim().slice(0, 100) : "Canvas LMS"

  if (!baseUrl) {
    return NextResponse.json({ error: "Canvas instance URL is required" }, { status: 400 })
  }
  if (!rawCookies) {
    return NextResponse.json({ error: "Cookies are required" }, { status: 400 })
  }

  // Store base_url and the full raw cookie string
  const credentialData = JSON.stringify({
    base_url: baseUrl,
    cookies: rawCookies,
  })

  const encrypted = encrypt(credentialData)

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
