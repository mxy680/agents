import { createAdminClient } from "@/lib/supabase/admin"
import { resolveAdminCredentials } from "@/lib/credentials"
import { runContainer } from "@/lib/container-runner"
import path from "path"
import os from "os"

interface JobDefinition {
  id: string
  slug: string
  prompt: string
  timeout_minutes: number
  agent_templates: { name: string } | { name: string }[]
}

/**
 * Execute a job for a user. Creates a job_runs row, spawns a container,
 * collects output, and updates the row on completion.
 */
export async function runJob(userId: string, jobDefinitionId: string): Promise<string> {
  const admin = createAdminClient()

  // Fetch job definition with template name
  const { data: jobDef, error: jobError } = await admin
    .from("job_definitions")
    .select("id, slug, prompt, timeout_minutes, agent_templates(name)")
    .eq("id", jobDefinitionId)
    .single()

  if (jobError || !jobDef) {
    throw new Error(`Job definition not found: ${jobDefinitionId}`)
  }

  const typedJobDef = jobDef as unknown as JobDefinition

  const templateName = Array.isArray(typedJobDef.agent_templates)
    ? typedJobDef.agent_templates[0]?.name
    : (typedJobDef.agent_templates as { name: string })?.name

  if (!templateName) {
    throw new Error(`No template found for job ${jobDefinitionId}`)
  }

  // Create job run row
  const { data: run, error: runError } = await admin
    .from("job_runs")
    .insert({
      job_definition_id: jobDefinitionId,
      user_id: userId,
      status: "running",
      started_at: new Date().toISOString(),
    })
    .select("id")
    .single()

  if (runError || !run) {
    throw new Error(`Failed to create job run: ${runError?.message}`)
  }

  const runId = (run as { id: string }).id
  const startTime = Date.now()

  try {
    // Resolve credentials
    const credEnv = await resolveAdminCredentials()

    const claudeToken = process.env.CLAUDE_CODE_OAUTH_TOKEN
    if (!claudeToken) {
      throw new Error("CLAUDE_CODE_OAUTH_TOKEN not configured")
    }

    // Build instance directory
    const instancePath = path.join(os.tmpdir(), "jobs", userId, typedJobDef.slug, String(Date.now()))

    // Copy agent template files
    const { mkdirSync, copyFileSync, existsSync } = await import("fs")
    mkdirSync(instancePath, { recursive: true })

    const agentsDir = path.join(process.cwd(), "..", "agents", templateName)
    for (const file of ["role.md", "CLAUDE.md"]) {
      const src = path.join(agentsDir, file)
      if (existsSync(src)) {
        copyFileSync(src, path.join(instancePath, file))
      }
    }

    const containerEnv: Record<string, string> = {
      CLAUDE_CODE_OAUTH_TOKEN: claudeToken,
      ...credEnv,
    }

    // Collect output text from container
    let outputText = ""
    let lastError = ""
    for await (const event of runContainer({
      instancePath,
      message: typedJobDef.prompt,
      env: containerEnv,
      timeoutMs: (typedJobDef.timeout_minutes ?? 10) * 60 * 1000,
    })) {
      if (event.event === "delta") {
        outputText += event.data
      } else if (event.event === "error") {
        lastError = event.data
        console.error(`[job-runner] Error for job ${typedJobDef.slug}:`, event.data)
      }
    }

    // If no output but there was an error, treat as failure
    if (!outputText && lastError) {
      throw new Error(lastError)
    }

    const durationMs = Date.now() - startTime

    // Update run as completed
    await admin
      .from("job_runs")
      .update({
        status: "completed",
        output_markdown: outputText,
        completed_at: new Date().toISOString(),
        duration_ms: durationMs,
      })
      .eq("id", runId)

    return runId
  } catch (e) {
    const durationMs = Date.now() - startTime
    const errorMsg = e instanceof Error ? e.message : String(e)

    await admin
      .from("job_runs")
      .update({
        status: "failed",
        error_message: errorMsg,
        completed_at: new Date().toISOString(),
        duration_ms: durationMs,
      })
      .eq("id", runId)

    throw e
  }
}
