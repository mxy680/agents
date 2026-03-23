import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { isAdmin } from "@/lib/admin"

export async function GET() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user || !isAdmin(user.email)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  const admin = createAdminClient()
  const { data: clients, error } = await admin
    .from("clients")
    .select("id, name, email, notes, active, created_at")
    .order("created_at", { ascending: false })

  if (error) {
    return NextResponse.json({ error: error.message }, { status: 500 })
  }

  return NextResponse.json({ clients })
}

export async function POST(request: NextRequest) {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user || !isAdmin(user.email)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  const body = await request.json()
  const { name, email, notes } = body as {
    name: string
    email?: string | null
    notes?: string | null
  }

  if (!name || typeof name !== "string" || !name.trim()) {
    return NextResponse.json({ error: "name is required" }, { status: 400 })
  }

  const admin = createAdminClient()
  const { data: client, error } = await admin
    .from("clients")
    .insert({
      name: name.trim(),
      email: email?.trim() || null,
      notes: notes?.trim() || null,
    })
    .select("id, name, email, notes, active, created_at")
    .single()

  if (error) {
    return NextResponse.json({ error: error.message }, { status: 500 })
  }

  return NextResponse.json({ client }, { status: 201 })
}
