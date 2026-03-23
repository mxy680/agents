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
  const proxyUrl = typeof body.proxy_url === "string" ? body.proxy_url.trim() : ""
  const label = typeof body.label === "string" ? body.label.trim().slice(0, 100) : "Zillow"

  // Validate proxy URL format if provided
  if (proxyUrl && !proxyUrl.match(/^(https?|socks[45]):\/\//)) {
    return NextResponse.json(
      { error: "Invalid proxy URL. Must start with http://, https://, socks4://, or socks5://" },
      { status: 400 }
    )
  }

  const credentials = JSON.stringify({
    proxy_url: proxyUrl,
    configured: "true",
  })
  const encrypted = encrypt(credentials)

  const admin = createAdminClient()
  const { error: dbError } = await admin
    .from("user_integrations")
    .upsert(
      {
        user_id: user.id,
        provider: "zillow",
        label,
        status: "active",
        credentials: `\\x${Buffer.from(encrypted).toString("hex")}`,
        updated_at: new Date().toISOString(),
      },
      { onConflict: "user_id,provider,label" }
    )

  if (dbError) {
    console.error("[zillow/save] DB error:", dbError)
    return NextResponse.json({ error: "Failed to save configuration" }, { status: 500 })
  }

  return NextResponse.json({ success: true })
}
