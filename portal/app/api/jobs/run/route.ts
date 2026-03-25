export const maxDuration = 300

import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { isAdmin } from "@/lib/admin"
import { spawn } from "child_process"
import path from "path"

// Repo root is one level above portal/
const REPO_ROOT = path.resolve(process.cwd(), "..")

export async function POST(request: NextRequest) {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user || !isAdmin(user.email)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  let agent: string
  let job: string
  try {
    const body = await request.json()
    agent = body.agent
    job = body.job
  } catch {
    return NextResponse.json({ error: "Invalid request body" }, { status: 400 })
  }

  if (!agent || !job) {
    return NextResponse.json({ error: "agent and job are required" }, { status: 400 })
  }

  const admin = createAdminClient()

  // Insert a pending run record
  const { data: run, error: insertError } = await admin
    .from("local_job_runs")
    .insert({
      agent_name: agent,
      job_slug: job,
      status: "pending",
    })
    .select("id")
    .single()

  if (insertError || !run) {
    console.error("[jobs/run] Failed to insert run:", insertError)
    return NextResponse.json({ error: "Failed to create run record" }, { status: 500 })
  }

  const runId = run.id

  // Fire and forget — don't await
  runPipeline(runId, agent, job).catch((err) => {
    console.error(`[jobs/run] Unhandled error for run ${runId}:`, err)
  })

  return NextResponse.json({ runId })
}

async function runPipeline(runId: string, agent: string, job: string) {
  const admin = createAdminClient()

  // Mark as running immediately
  await admin
    .from("local_job_runs")
    .update({ status: "running", started_at: new Date().toISOString() })
    .eq("id", runId)

  const resolveCredsScript = path.join(REPO_ROOT, "agents", agent, "resolve-creds.mjs")
  const pipelineScript = path.join(REPO_ROOT, "agents", agent, "scripts", "run_pipeline.sh")
  const binPath = path.join(REPO_ROOT, "bin")

  const command = `
    doppler run --project agents --config dev -- node "${resolveCredsScript}" > /tmp/job_creds_${runId}.sh && \
    doppler run --project agents --config dev -- bash -c "
      source /tmp/job_creds_${runId}.sh
      export PATH='${binPath}':$PATH
      bash '${pipelineScript}'
    "
  `

  const child = spawn("bash", ["-c", command], {
    env: { ...process.env },
    cwd: REPO_ROOT,
  })

  let logBuffer = ""
  let lastFlush = Date.now()
  const FLUSH_INTERVAL_MS = 500

  async function flushLog(force = false) {
    if (!force && Date.now() - lastFlush < FLUSH_INTERVAL_MS) return
    if (!logBuffer) return
    const chunk = logBuffer
    logBuffer = ""
    lastFlush = Date.now()
    // Append the chunk to the log column
    const { data: current } = await admin
      .from("local_job_runs")
      .select("log")
      .eq("id", runId)
      .single()
    const existingLog = current?.log ?? ""
    await admin
      .from("local_job_runs")
      .update({ log: existingLog + chunk })
      .eq("id", runId)
  }

  child.stdout.on("data", async (data: Buffer) => {
    logBuffer += data.toString()
    await flushLog()
  })

  child.stderr.on("data", async (data: Buffer) => {
    logBuffer += data.toString()
    await flushLog()
  })

  await new Promise<void>((resolve) => {
    child.on("close", async (code) => {
      // Flush any remaining buffered output
      await flushLog(true)

      // Parse the last lines of the full log for deliverable JSON
      const { data: finalRun } = await admin
        .from("local_job_runs")
        .select("log")
        .eq("id", runId)
        .single()

      const fullLog = finalRun?.log ?? ""
      const deliverables = parseDeliverables(fullLog)

      const status = code === 0 ? "completed" : "failed"
      await admin
        .from("local_job_runs")
        .update({
          status,
          completed_at: new Date().toISOString(),
          deliverables,
        })
        .eq("id", runId)

      // Clean up temp creds file
      try {
        const { execSync } = await import("child_process")
        execSync(`rm -f /tmp/job_creds_${runId}.sh`)
      } catch {
        // Ignore cleanup errors
      }

      resolve()
    })
  })
}

/**
 * Parse deliverable links from log output.
 * Looks for JSON objects containing sheet_url or pdf_url keys in the last 50 lines.
 */
function parseDeliverables(log: string): Record<string, string> {
  const lines = log.split("\n")
  const tail = lines.slice(-50).join("\n")

  // Try to find JSON objects with deliverable URLs
  const jsonPattern = /\{[^{}]*(?:sheet_url|pdf_url|spreadsheet_url|report_url)[^{}]*\}/g
  const matches = tail.match(jsonPattern)

  if (matches) {
    for (const match of matches.reverse()) {
      try {
        const parsed = JSON.parse(match)
        if (parsed && typeof parsed === "object") {
          return parsed as Record<string, string>
        }
      } catch {
        continue
      }
    }
  }

  // Also look for Google Drive URLs in the log
  const driveUrlPattern = /https:\/\/docs\.google\.com\/[^\s"']+/g
  const driveUrls = tail.match(driveUrlPattern)
  if (driveUrls && driveUrls.length > 0) {
    const deliverables: Record<string, string> = {}
    for (const url of driveUrls) {
      if (url.includes("/spreadsheets/")) {
        deliverables.sheet_url = url
      } else if (url.includes("/document/") || url.includes("drive.google.com")) {
        deliverables.drive_url = url
      }
    }
    // Also look for PDF uploads
    const pdfLine = tail.split("\n").find((l) => l.includes(".pdf") && l.includes("drive.google.com"))
    if (pdfLine) {
      const pdfUrl = pdfLine.match(/https:\/\/[^\s"']+/)?.[0]
      if (pdfUrl) deliverables.pdf_url = pdfUrl
    }
    if (Object.keys(deliverables).length > 0) return deliverables
  }

  return {}
}
