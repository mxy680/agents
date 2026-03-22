import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { isAdmin } from "@/lib/admin"

async function getTemplateByName(admin: ReturnType<typeof createAdminClient>, name: string) {
  const { data } = await admin
    .from("agent_templates")
    .select("id")
    .eq("name", name)
    .single()
  return data
}

export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ name: string }> }
) {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user || !isAdmin(user.email)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  const { name } = await params
  const admin = createAdminClient()
  const template = await getTemplateByName(admin, name)

  if (!template) {
    return NextResponse.json({ error: "Agent not found" }, { status: 404 })
  }

  const { data: assignments, error } = await admin
    .from("client_agents")
    .select("id, client_id, notes, created_at, clients(id, name, email, active)")
    .eq("template_id", template.id)
    .order("created_at", { ascending: false })

  if (error) {
    return NextResponse.json({ error: error.message }, { status: 500 })
  }

  return NextResponse.json({ assignments: assignments ?? [] })
}

export async function POST(
  request: NextRequest,
  { params }: { params: Promise<{ name: string }> }
) {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user || !isAdmin(user.email)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  const { name } = await params
  const body = await request.json()
  const { client_id, notes } = body as { client_id: string; notes?: string }

  if (!client_id) {
    return NextResponse.json({ error: "client_id is required" }, { status: 400 })
  }

  const admin = createAdminClient()
  const template = await getTemplateByName(admin, name)

  if (!template) {
    return NextResponse.json({ error: "Agent not found" }, { status: 404 })
  }

  const { data: assignment, error } = await admin
    .from("client_agents")
    .insert({
      client_id,
      template_id: template.id,
      notes: notes?.trim() || null,
    })
    .select("id, client_id, notes, created_at")
    .single()

  if (error) {
    if (error.code === "23505") {
      return NextResponse.json({ error: "Client already assigned to this agent" }, { status: 409 })
    }
    return NextResponse.json({ error: error.message }, { status: 500 })
  }

  return NextResponse.json({ assignment }, { status: 201 })
}

export async function DELETE(
  request: NextRequest,
  { params }: { params: Promise<{ name: string }> }
) {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user || !isAdmin(user.email)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  const { name } = await params
  const body = await request.json()
  const { client_id } = body as { client_id: string }

  if (!client_id) {
    return NextResponse.json({ error: "client_id is required" }, { status: 400 })
  }

  const admin = createAdminClient()
  const template = await getTemplateByName(admin, name)

  if (!template) {
    return NextResponse.json({ error: "Agent not found" }, { status: 404 })
  }

  const { error } = await admin
    .from("client_agents")
    .delete()
    .eq("template_id", template.id)
    .eq("client_id", client_id)

  if (error) {
    return NextResponse.json({ error: error.message }, { status: 500 })
  }

  return NextResponse.json({ ok: true })
}
