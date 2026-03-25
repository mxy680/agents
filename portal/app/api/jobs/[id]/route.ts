import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { isAdmin } from "@/lib/admin"

export async function GET(
  _request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params

  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user || !isAdmin(user.email)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  const admin = createAdminClient()

  const { data: run, error } = await admin
    .from("local_job_runs")
    .select("id, status, started_at, completed_at, deliverables, log")
    .eq("id", id)
    .single()

  if (error || !run) {
    return NextResponse.json({ error: "Not found" }, { status: 404 })
  }

  return NextResponse.json({
    id: run.id,
    status: run.status,
    started_at: run.started_at,
    completed_at: run.completed_at,
    deliverables: run.deliverables ?? {},
    log_length: (run.log ?? "").length,
  })
}
