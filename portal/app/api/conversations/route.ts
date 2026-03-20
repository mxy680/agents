import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"

export async function GET(request: NextRequest) {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  const agentName = request.nextUrl.searchParams.get("agent_name")

  let query = supabase
    .from("conversations")
    .select("*")
    .eq("user_id", user.id)
    .order("updated_at", { ascending: false })

  if (agentName) {
    query = query.eq("agent_name", agentName)
  }

  const { data, error } = await query

  if (error) {
    console.error("[conversations] GET error:", error.message)
    return NextResponse.json({ error: "Failed to fetch conversations" }, { status: 500 })
  }

  return NextResponse.json({ conversations: data ?? [] })
}

export async function POST(request: NextRequest) {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  let agent_name: string
  try {
    const body = await request.json()
    agent_name = body.agent_name
  } catch {
    return NextResponse.json({ error: "Invalid request body" }, { status: 400 })
  }

  if (!agent_name || typeof agent_name !== "string") {
    return NextResponse.json({ error: "agent_name is required" }, { status: 400 })
  }

  const { data, error } = await supabase
    .from("conversations")
    .insert({ user_id: user.id, agent_name })
    .select("id")
    .single()

  if (error) {
    console.error("[conversations] POST error:", error.message)
    return NextResponse.json({ error: "Failed to create conversation" }, { status: 500 })
  }

  return NextResponse.json({ id: data.id }, { status: 201 })
}
