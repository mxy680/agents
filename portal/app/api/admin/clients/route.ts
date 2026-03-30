import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { isAdmin } from "@/lib/admin"

export async function POST(request: NextRequest) {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user || !isAdmin(user.email)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  const body = await request.json()
  const { code, client_name, agent_names } = body as {
    code: string
    client_name: string
    agent_names: string[]
  }

  if (!code?.trim() || !client_name?.trim() || !agent_names?.length) {
    return NextResponse.json({ error: "code, client_name, and agent_names required" }, { status: 400 })
  }

  const admin = createAdminClient()
  const { data: client, error } = await admin
    .from("client_access")
    .insert({
      code: code.trim(),
      client_name: client_name.trim(),
      agent_name: agent_names[0],
      agent_names,
    })
    .select("id, code, client_name, agent_name, agent_names, active, created_at")
    .single()

  if (error) {
    if (error.code === "23505") {
      return NextResponse.json({ error: "Access code already exists" }, { status: 409 })
    }
    return NextResponse.json({ error: error.message }, { status: 500 })
  }

  return NextResponse.json({ client }, { status: 201 })
}
