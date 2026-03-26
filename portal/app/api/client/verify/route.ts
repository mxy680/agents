import { NextRequest, NextResponse } from "next/server"
import { createAdminClient } from "@/lib/supabase/admin"

export async function POST(request: NextRequest) {
  let code: string
  try {
    const body = await request.json()
    code = body.code
  } catch {
    return NextResponse.json({ error: "Invalid request" }, { status: 400 })
  }

  if (!code || typeof code !== "string") {
    return NextResponse.json({ error: "Code is required" }, { status: 400 })
  }

  const admin = createAdminClient()
  const { data, error } = await admin
    .from("client_access")
    .select("id, client_name, agent_name, active")
    .eq("code", code.trim())
    .eq("active", true)
    .single()

  if (error || !data) {
    return NextResponse.json({ error: "Invalid access code" }, { status: 401 })
  }

  return NextResponse.json({
    ok: true,
    clientName: data.client_name,
    agentName: data.agent_name,
  })
}
