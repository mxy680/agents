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
  const projectUrl = typeof body.project_url === "string" ? body.project_url.trim() : ""
  const label = typeof body.label === "string" ? body.label.trim().slice(0, 100) : "Framer Project"

  if (!apiKey) {
    return NextResponse.json({ error: "API key is required" }, { status: 400 })
  }
  if (!projectUrl) {
    return NextResponse.json({ error: "Project URL is required" }, { status: 400 })
  }

  // Validate project URL format
  if (!projectUrl.startsWith("https://framer.com/projects/")) {
    return NextResponse.json(
      { error: "Invalid project URL. Must start with https://framer.com/projects/" },
      { status: 400 }
    )
  }

  const credentials = JSON.stringify({ api_key: apiKey, project_url: projectUrl })
  const encrypted = encrypt(credentials)

  const admin = createAdminClient()
  const { error: dbError } = await admin
    .from("user_integrations")
    .upsert(
      {
        user_id: user.id,
        provider: "framer",
        label,
        status: "active",
        credentials: `\\x${Buffer.from(encrypted).toString("hex")}`,
        updated_at: new Date().toISOString(),
      },
      { onConflict: "user_id,provider,label" }
    )

  if (dbError) {
    console.error("[framer/save] DB error:", dbError)
    return NextResponse.json({ error: "Failed to save credentials" }, { status: 500 })
  }

  return NextResponse.json({ success: true })
}
