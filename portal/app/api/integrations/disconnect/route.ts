import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { checkOrigin } from "@/lib/csrf"

export async function POST(request: NextRequest) {
  const csrfError = checkOrigin(request)
  if (csrfError) return csrfError

  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  let id: string
  try {
    const body = await request.json()
    id = body.id
  } catch {
    return NextResponse.json({ error: "Invalid request body" }, { status: 400 })
  }

  if (!id || typeof id !== "string") {
    return NextResponse.json({ error: "Missing integration id" }, { status: 400 })
  }

  const admin = createAdminClient()
  const { error } = await admin
    .from("user_integrations")
    .delete()
    .eq("id", id)
    .eq("user_id", user.id)

  if (error) {
    return NextResponse.json({ error: error.message }, { status: 500 })
  }

  return NextResponse.json({ ok: true })
}
