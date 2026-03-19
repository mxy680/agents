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

interface UserAgent {
  user_id: string
  status: string
  template_id: string
}

interface JobRun {
  id: string
  status: string
}

export async function GET(req: NextRequest) {
  // Validate CRON_SECRET header to prevent external abuse
  const cronSecret = process.env.CRON_SECRET
  if (cronSecret) {
    const authHeader = req.headers.get("x-cron-secret")
    if (authHeader !== cronSecret) {
      return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
    }
  }

  const admin = createAdminClient()
  const now = new Date()

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

  const triggered: Array<{ jobId: string; jobName: string; userId: string }> = []
  const skipped: Array<{ jobId: string; jobName: string; reason: string }> = []

  for (const job of definitions) {
    // Check if job is due
    if (!isDue(job.schedule, now)) {
      skipped.push({ jobId: job.id, jobName: job.display_name, reason: "not due" })
      continue
    }

    // Find all users with approved access to this template
    const { data: userAgents, error: userAgentsError } = await admin
      .from("user_agents")
      .select("user_id, status, template_id")
      .eq("template_id", job.template_id)
      .eq("status", "approved")

    if (userAgentsError) {
      console.error(`[cron] Failed to fetch user_agents for job ${job.id}:`, userAgentsError.message)
      skipped.push({ jobId: job.id, jobName: job.display_name, reason: "user_agents fetch error" })
      continue
    }

    const approvedUsers = (userAgents ?? []) as UserAgent[]

    for (const ua of approvedUsers) {
      // Check for existing pending/running runs to prevent duplicates
      const { data: existingRuns } = await admin
        .from("job_runs")
        .select("id, status")
        .eq("job_definition_id", job.id)
        .eq("user_id", ua.user_id)
        .in("status", ["pending", "running"])
        .limit(1)

      const runs = (existingRuns ?? []) as JobRun[]

      if (runs.length > 0) {
        skipped.push({
          jobId: job.id,
          jobName: job.display_name,
          reason: `user ${ua.user_id} already has a pending/running run`,
        })
        continue
      }

      // Fire and forget — do not await
      runJob(ua.user_id, job.id).catch((err: unknown) => {
        console.error(
          `[cron] runJob failed for job ${job.id} user ${ua.user_id}:`,
          err instanceof Error ? err.message : String(err)
        )
      })

      triggered.push({ jobId: job.id, jobName: job.display_name, userId: ua.user_id })
    }
  }

  return NextResponse.json({
    ok: true,
    triggeredAt: now.toISOString(),
    triggered,
    skipped,
  })
}
