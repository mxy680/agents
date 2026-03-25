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

  // Kill the wrapper process if still running
  if (run.status === "running" || run.status === "pending") {
    const wrapperScript = `/tmp/job_wrapper_${id}.sh`
    try {
      // Find all processes spawned from the wrapper script and kill the process group
      execSync(`pkill -f "job_wrapper_${id}" 2>/dev/null || true`)
      // Also kill any child doppler/node/python processes for this run
      execSync(`pkill -f "job_creds_${id}" 2>/dev/null || true`)
      execSync(`pkill -f "job_log_${id}" 2>/dev/null || true`)
    } catch {
      // Process may already be dead — that's fine
    }

    // Clean up temp files
    try {
      execSync(`rm -f /tmp/job_wrapper_${id}.sh /tmp/job_creds_${id}.sh /tmp/job_log_${id}.txt 2>/dev/null || true`)
    } catch {
      // Ignore cleanup errors
    }
  }

  // Mark as cancelled in DB
  await admin
    .from("local_job_runs")
    .update({
      status: "failed",
      completed_at: new Date().toISOString(),
      log: (run.status === "running" || run.status === "pending")
        ? undefined  // preserve existing log
        : undefined,
    })
    .eq("id", id)

  // Append cancellation note to log
  const { data: current } = await admin
    .from("local_job_runs")
    .select("log")
    .eq("id", id)
    .single()

  await admin
    .from("local_job_runs")
    .update({
      status: "failed",
      completed_at: new Date().toISOString(),
      log: (current?.log ?? "") + "\n\n[Cancelled by user]",
    })
    .eq("id", id)

  return NextResponse.json({ cancelled: true })
}
