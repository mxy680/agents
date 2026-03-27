import { NextRequest, NextResponse } from "next/server"
import { createAdminClient } from "@/lib/supabase/admin"

/**
 * GET /api/client/conversations/[id]?code=XXX
 *
 * Load a conversation with its messages (validates access code).
 */
export async function GET(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params
  const code = request.nextUrl.searchParams.get("code")

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

  // Fetch conversation
  const { data: conversation } = await admin
    .from("conversations")
    .select("id, agent_name, title, created_at")
    .eq("id", id)
    .single()

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
