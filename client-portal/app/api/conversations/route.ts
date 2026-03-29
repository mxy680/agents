import { NextRequest, NextResponse } from "next/server"
import { createAdminClient } from "@/lib/supabase/admin"

/**
 * GET /api/conversations?agent=YYY
 *
 * List conversations for a client access code (from session cookie) + agent.
 */
export async function GET(request: NextRequest) {
  const code = request.cookies.get("engagent_session")?.value
  const agent = request.nextUrl.searchParams.get("agent")

  if (!code) {
    return NextResponse.json({ error: "code required" }, { status: 400 })
  }

  const admin = createAdminClient()

  // Validate code
  const { data: access } = await admin
    .from("client_access")
    .select("id")
    .eq("code", code)
    .eq("active", true)
    .single()

  if (!access) {
    return NextResponse.json({ error: "Invalid code" }, { status: 401 })
  }

  // Fetch conversations — match by agent_name and a marker in the conversation
  // Client conversations use user_id = 00000000-... as a placeholder
  let query = admin
    .from("conversations")
    .select("id, title, agent_name, updated_at, created_at")
    .eq("user_id", "00000000-0000-0000-0000-000000000000")
    .order("updated_at", { ascending: false })
    .limit(50)

  if (agent) {
    query = query.eq("agent_name", agent)
  }

  const { data: conversations } = await query

  return NextResponse.json({ conversations: conversations ?? [] })
}

/**
 * DELETE /api/conversations?id=XXX
 *
 * Delete a conversation (validates session cookie first).
 */
export async function DELETE(request: NextRequest) {
  const id = request.nextUrl.searchParams.get("id")
  const code = request.cookies.get("engagent_session")?.value

  if (!id || !code) {
    return NextResponse.json({ error: "id and code required" }, { status: 400 })
  }

  const admin = createAdminClient()

  // Validate code
  const { data: access } = await admin
    .from("client_access")
    .select("id")
    .eq("code", code)
    .eq("active", true)
    .single()

  if (!access) {
    return NextResponse.json({ error: "Invalid code" }, { status: 401 })
  }

  // Delete messages first (cascade may not be set up)
  await admin.from("conversation_messages").delete().eq("conversation_id", id)
  await admin.from("conversations").delete().eq("id", id)

  return NextResponse.json({ ok: true })
}
