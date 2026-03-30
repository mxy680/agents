import { NextRequest, NextResponse } from "next/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { verifySession } from "@/lib/session"

/**
 * GET /api/conversations/[id]
 *
 * Load a conversation with its messages (validates session cookie).
 */
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params
  const cookieValue = request.cookies.get("engagent_session")?.value
  const code = cookieValue ? verifySession(cookieValue) : null

  if (!code) {
    return NextResponse.json({ error: "code required" }, { status: 401 })
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

  // Fetch conversation scoped to this client (client_access_id must match when set)
  const { data: conversation } = await admin
    .from("conversations")
    .select("id, agent_name, title, created_at, client_access_id")
    .eq("id", id)
    .single()

  // Enforce ownership: if client_access_id is set, it must match this client
  if (conversation?.client_access_id && conversation.client_access_id !== access.id) {
    return NextResponse.json({ error: "Forbidden" }, { status: 403 })
  }

  if (!conversation) {
    return NextResponse.json({ error: "Conversation not found" }, { status: 404 })
  }

  // Fetch messages
  const { data: messages } = await admin
    .from("conversation_messages")
    .select("id, role, blocks, created_at")
    .eq("conversation_id", id)
    .order("created_at", { ascending: true })

  return NextResponse.json({ conversation, messages: messages ?? [] })
}
