import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"

export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  const { data: conversation, error: convError } = await supabase
    .from("conversations")
    .select("*")
    .eq("id", id)
    .eq("user_id", user.id)
    .single()

  if (convError || !conversation) {
    return NextResponse.json({ error: "Conversation not found" }, { status: 404 })
  }

  const { data: messages, error: msgError } = await supabase
    .from("conversation_messages")
    .select("*")
    .eq("conversation_id", id)
    .order("created_at", { ascending: true })

  if (msgError) {
    console.error("[conversations/id] GET messages error:", msgError.message)
    return NextResponse.json({ error: "Failed to fetch messages" }, { status: 500 })
  }

  return NextResponse.json({ conversation, messages: messages ?? [] })
}

export async function PATCH(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  // Verify ownership
  const { data: existing } = await supabase
    .from("conversations")
    .select("id")
    .eq("id", id)
    .eq("user_id", user.id)
    .single()

  if (!existing) {
    return NextResponse.json({ error: "Conversation not found" }, { status: 404 })
  }

  let body: { title?: string; starred?: boolean }
  try {
    body = await request.json()
  } catch {
    return NextResponse.json({ error: "Invalid request body" }, { status: 400 })
  }

  const updates: Record<string, unknown> = {}
  if (typeof body.title === "string") updates.title = body.title
  if (typeof body.starred === "boolean") updates.starred = body.starred

  if (Object.keys(updates).length === 0) {
    return NextResponse.json({ error: "No valid fields to update" }, { status: 400 })
  }

  const { data, error } = await supabase
    .from("conversations")
    .update(updates)
    .eq("id", id)
    .eq("user_id", user.id)
    .select("*")
    .single()

  if (error) {
    console.error("[conversations/id] PATCH error:", error.message)
    return NextResponse.json({ error: "Failed to update conversation" }, { status: 500 })
  }

  return NextResponse.json({ conversation: data })
}

export async function DELETE(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  // Verify ownership
  const { data: existing } = await supabase
    .from("conversations")
    .select("id")
    .eq("id", id)
    .eq("user_id", user.id)
    .single()

  if (!existing) {
    return NextResponse.json({ error: "Conversation not found" }, { status: 404 })
  }

  // Use admin client to delete messages first (bypass RLS for service-role actions)
  const admin = createAdminClient()
  const { error: msgDeleteError } = await admin
    .from("conversation_messages")
    .delete()
    .eq("conversation_id", id)

  if (msgDeleteError) {
    console.error("[conversations/id] DELETE messages error:", msgDeleteError.message)
    return NextResponse.json({ error: "Failed to delete messages" }, { status: 500 })
  }

  const { error } = await supabase
    .from("conversations")
    .delete()
    .eq("id", id)
    .eq("user_id", user.id)

  if (error) {
    console.error("[conversations/id] DELETE error:", error.message)
    return NextResponse.json({ error: "Failed to delete conversation" }, { status: 500 })
  }

  return NextResponse.json({ ok: true })
}
