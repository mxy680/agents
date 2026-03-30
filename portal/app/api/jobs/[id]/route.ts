import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { isAdmin } from "@/lib/admin"
import { execSync } from "child_process"

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
    .select("id, agent_name, job_slug, status, started_at, completed_at, deliverables, log")
    .eq("id", id)
    .single()

  if (error || !run) {
    return NextResponse.json({ error: "Not found" }, { status: 404 })
  }

  return NextResponse.json({
    id: run.id,
    agent_name: run.agent_name,
    job_slug: run.job_slug,
    status: run.status,
    started_at: run.started_at,
    completed_at: run.completed_at,
    deliverables: run.deliverables ?? {},
    log: run.log ?? "",
    log_length: (run.log ?? "").length,
  })
}

export async function DELETE(
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
    .select("id, status")
    .eq("id", id)
    .single()

  if (error || !run) {
    return NextResponse.json({ error: "Not found" }, { status: 404 })
  }

  // If running/pending, kill the process first
  if (run.status === "running" || run.status === "pending") {
    try {
      execSync(`pkill -f "job_wrapper_${id}" 2>/dev/null || true`)
      execSync(`pkill -f "job_creds_${id}" 2>/dev/null || true`)
      execSync(`pkill -f "job_log_${id}" 2>/dev/null || true`)
    } catch {
      // Process may already be dead
    }

    try {
      execSync(`rm -f /tmp/job_wrapper_${id}.sh /tmp/job_creds_${id}.sh /tmp/job_log_${id}.txt /tmp/job_session_${id}.json 2>/dev/null || true`)
    } catch {
      // Ignore cleanup errors
    }
  }

  // Delete the run record
  await admin
    .from("local_job_runs")
    .delete()
    .eq("id", id)

  return NextResponse.json({ deleted: true })
}
