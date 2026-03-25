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
  const apiKey = typeof body.api_key === "string" ? body.api_key.trim() : ""
  const label = typeof body.label === "string" ? body.label.trim().slice(0, 100) : "Yelp"

  if (!apiKey) {
    return NextResponse.json({ error: "API key is required" }, { status: 400 })
  }

  const credentials = JSON.stringify({ api_key: apiKey })
  const encrypted = encrypt(credentials)

  const admin = createAdminClient()
  const { error: dbError } = await admin
    .from("user_integrations")
    .upsert(
      {
        user_id: user.id,
        provider: "yelp",
        label,
        status: "active",
        credentials: `\\x${Buffer.from(encrypted).toString("hex")}`,
        updated_at: new Date().toISOString(),
      },
      { onConflict: "user_id,provider,label" }
    )

  if (dbError) {
    console.error("[yelp/save] DB error:", dbError)
    return NextResponse.json({ error: "Failed to save credentials" }, { status: 500 })
  }

  return NextResponse.json({ success: true })
}
