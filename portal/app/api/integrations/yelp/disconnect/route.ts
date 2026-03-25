import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
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
  const integrationId = typeof body.id === "string" ? body.id : ""

  if (!integrationId) {
    return NextResponse.json({ error: "Integration ID is required" }, { status: 400 })
  }

  const admin = createAdminClient()
  const { error: dbError } = await admin
    .from("user_integrations")
    .delete()
    .eq("id", integrationId)
    .eq("user_id", user.id)
    .eq("provider", "yelp")

  if (dbError) {
    console.error("[yelp/disconnect] DB error:", dbError)
    return NextResponse.json({ error: "Failed to disconnect" }, { status: 500 })
  }

  return NextResponse.json({ success: true })
}
