import { NextRequest, NextResponse } from "next/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { isDue } from "@/lib/cron"
import { runJob } from "@/lib/job-runner"

interface JobDefinition {
  id: string
  template_id: string
  slug: string
  schedule: string
  display_name: string
  enabled: boolean
}

interface JobRun {
  id: string
  status: string
}

export async function GET(req: NextRequest) {
  // Validate CRON_SECRET header to prevent external abuse
  const cronSecret = process.env.CRON_SECRET
  if (!cronSecret) {
    console.error("[cron] CRON_SECRET is not configured")
    return NextResponse.json({ error: "Server misconfigured" }, { status: 500 })
  }
  const authHeader = req.headers.get("x-cron-secret")
  if (authHeader !== cronSecret) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  const admin = createAdminClient()
  const now = new Date()

  // Single-tenant: use admin user ID for all job runs
  const adminUserId = process.env.ADMIN_USER_ID
  if (!adminUserId) {
    console.error("[cron] ADMIN_USER_ID is not configured")
    return NextResponse.json({ error: "ADMIN_USER_ID not configured" }, { status: 500 })
  }

  // Fetch all enabled job definitions
  const { data: jobDefs, error: jobDefsError } = await admin
    .from("job_definitions")
    .select("id, template_id, slug, schedule, display_name, enabled")
    .eq("enabled", true)

  if (jobDefsError) {
    console.error("[cron] Failed to fetch job definitions:", jobDefsError.message)
    return NextResponse.json({ error: "Failed to fetch job definitions" }, { status: 500 })
  }

  const definitions = (jobDefs ?? []) as JobDefinition[]

  const triggered: Array<{ jobId: string; jobName: string }> = []
  const skipped: Array<{ jobId: string; jobName: string; reason: string }> = []

  for (const job of definitions) {
    // Check if job is due
    if (!isDue(job.schedule, now)) {
      skipped.push({ jobId: job.id, jobName: job.display_name, reason: "not due" })
      continue
    }

    // Check for existing pending/running runs to prevent duplicates
    const { data: existingRuns } = await admin
      .from("job_runs")
      .select("id, status")
      .eq("job_definition_id", job.id)
      .in("status", ["pending", "running"])
      .limit(1)

    const runs = (existingRuns ?? []) as JobRun[]

    if (runs.length > 0) {
      skipped.push({
        jobId: job.id,
        jobName: job.display_name,
        reason: "already has a pending/running run",
      })
      continue
    }

    // Fire and forget — run with admin credentials
    runJob(adminUserId, job.id).catch((err: unknown) => {
      console.error(
        `[cron] runJob failed for job ${job.id}:`,
        err instanceof Error ? err.message : String(err)
      )
    })

    triggered.push({ jobId: job.id, jobName: job.display_name })
  }

  return NextResponse.json({
    ok: true,
    triggeredAt: now.toISOString(),
    triggered,
    skipped,
  })
}
