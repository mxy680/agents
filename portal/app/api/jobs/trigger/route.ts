import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { runJob } from "@/lib/job-runner"

export async function POST(request: NextRequest) {
  // Authenticate user
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  // Parse body
  let jobDefinitionId: string
  try {
    const body = await request.json()
    jobDefinitionId = body.jobDefinitionId
  } catch {
    return NextResponse.json({ error: "Invalid request body" }, { status: 400 })
  }

  if (!jobDefinitionId || typeof jobDefinitionId !== "string") {
    return NextResponse.json({ error: "jobDefinitionId is required" }, { status: 400 })
  }

  const admin = createAdminClient()

  // Validate job definition exists and get template_id
  const { data: jobDef, error: jobError } = await admin
    .from("job_definitions")
    .select("id, template_id, enabled")
    .eq("id", jobDefinitionId)
    .single()

  if (jobError || !jobDef) {
    return NextResponse.json({ error: "Job definition not found" }, { status: 404 })
  }

  if (!jobDef.enabled) {
    return NextResponse.json({ error: "Job is disabled" }, { status: 400 })
  }

  // Validate user has approved access to this agent
  const { data: userAgent } = await admin
    .from("user_agents")
    .select("status")
    .eq("user_id", user.id)
    .eq("template_id", jobDef.template_id)
    .single()

  if (!userAgent || userAgent.status !== "approved") {
    return NextResponse.json({ error: "Access denied" }, { status: 403 })
  }

  // Check for existing pending/running runs (prevent double-trigger)
  const { data: existingRuns } = await admin
    .from("job_runs")
    .select("id, status")
    .eq("job_definition_id", jobDefinitionId)
    .eq("user_id", user.id)
    .in("status", ["pending", "running"])

  if (existingRuns && existingRuns.length > 0) {
    return NextResponse.json({ error: "A run is already in progress" }, { status: 409 })
  }

  // Fire and forget - don't await
  runJob(user.id, jobDefinitionId).catch((e) => {
    console.error(`[jobs/trigger] Job ${jobDefinitionId} failed for user ${user.id}:`, e)
  })

  return NextResponse.json({ triggered: true, status: "running" })
}
