export const maxDuration = 30

import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { isAdmin } from "@/lib/admin"
import { spawn } from "child_process"
import { writeFileSync } from "fs"
import path from "path"

// Resolve repo root — same logic as local-runner.ts
const REPO_ROOT = (() => {
  const fromCwd = path.resolve(process.cwd(), "..")
  try {
    const { statSync } = require("fs")
    statSync(path.join(fromCwd, "agents"))
    return fromCwd // local dev
  } catch {
    return process.cwd() // Docker/Fly (/app)
  }
})()

// Detect if running in Docker/Fly (no Doppler CLI, secrets already in env)
const IS_CLOUD = process.env.NODE_ENV === "production" || !process.env.DOPPLER_TOKEN

// Allowlist of agents that can be triggered via this route
const ALLOWED_AGENTS = ["real-estate"]

// Map agent+job to the script path relative to agents/{agent}/scripts/
const JOB_SCRIPT_MAP: Record<string, string> = {
  "real-estate:weekly-scan": "run_pipeline.sh",
  "real-estate:off-market-scan": "off-market/run_pipeline.sh",
  "real-estate:probate-scan": "probate/run_scan.sh",
  "real-estate:llc-scan": "llc/run_scan.sh",
  "real-estate:wide-lot-scan": "wide-lot/run_scan.sh",
}

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

  // #11 Allowlist validation — prevent path injection
  if (!ALLOWED_AGENTS.includes(agent)) {
    return NextResponse.json({ error: `Unknown agent: ${agent}` }, { status: 400 })
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

  // Mark as running immediately (before spawning so status is visible right away)
  await admin
    .from("local_job_runs")
    .update({ status: "running", started_at: new Date().toISOString() })
    .eq("id", runId)

  // #1 Write a wrapper script that runs detached from the Next.js process.
  // The wrapper: resolves creds, runs the pipeline, and updates DB status on completion.
  // A trap ensures the creds file is always cleaned up.
  // resolve-creds.mjs lives in real-estate but is shared by all agents
  const resolveCredsScript = path.join(REPO_ROOT, "agents", "real-estate", "resolve-creds.mjs")

  // Look up the script path from the job map
  const scriptRelPath = JOB_SCRIPT_MAP[`${agent}:${job}`] ?? "run_pipeline.sh"
  const pipelineScript = path.join(REPO_ROOT, "agents", agent, "scripts", scriptRelPath)
  // On cloud, integrations binary is at /usr/local/bin (from Dockerfile)
  // On local, it's at REPO_ROOT/bin
  const binPath = IS_CLOUD ? "/usr/local/bin" : path.join(REPO_ROOT, "bin")
  const credsFile = `/tmp/job_creds_${runId}.sh`
  const logFile = `/tmp/job_log_${runId}.txt`
  const wrapperScript = `/tmp/job_wrapper_${runId}.sh`

  const portalNodeModules = IS_CLOUD
    ? path.join(REPO_ROOT, "node_modules")
    : path.join(REPO_ROOT, "portal", "node_modules")
  const dopplerToken = process.env.DOPPLER_SERVICE_TOKEN || process.env.DOPPLER_TOKEN || ""

  // On cloud: secrets are already in env (Doppler CMD wrapper), no need for doppler run
  // On local: use doppler run to inject secrets
  const dopplerPrefix = IS_CLOUD ? "" : "doppler run --project agents --config dev -- "

  const wrapperContents = `#!/bin/bash
set -uo pipefail
CREDS_FILE="${credsFile}"
LOG_FILE="${logFile}"
RUN_ID="${runId}"
export NODE_PATH="${portalNodeModules}:\${NODE_PATH:-}"
${dopplerToken ? `export DOPPLER_TOKEN="${dopplerToken}"` : ""}

# Always clean up creds file on exit
trap 'rm -f "$CREDS_FILE"' EXIT

# Resolve credentials
${dopplerPrefix}node "${resolveCredsScript}" > "$CREDS_FILE" 2>>"$LOG_FILE"
if [ $? -ne 0 ] || [ ! -s "$CREDS_FILE" ]; then
  echo "[wrapper] ERROR: credential resolution failed" >> "$LOG_FILE"
  exit 1
fi

# Source creds and run pipeline
export PATH="${binPath}:$PATH"
source "$CREDS_FILE"
bash "${pipelineScript}" >> "$LOG_FILE" 2>&1
EXIT_CODE=$?

echo "__EXIT_CODE__:$EXIT_CODE" >> "$LOG_FILE"
exit $EXIT_CODE
`

  writeFileSync(wrapperScript, wrapperContents, { mode: 0o755 })

  // #1 Spawn the wrapper fully detached so it outlives the Next.js request/process
  const child = spawn("bash", [wrapperScript], {
    detached: true,
    stdio: "ignore",
    env: { ...process.env },
    cwd: REPO_ROOT,
  })
  child.unref()

  // #5 Log flushing: accumulate chunks in memory, flush to DB every 2 seconds
  // This runs as a background task independent of the HTTP response.
  startLogTailer(runId, logFile, admin)

  return NextResponse.json({ runId })
}

/**
 * Polls the log file periodically and flushes new content to the DB.
 * Also watches for the __EXIT_CODE__ sentinel to finalize the run.
 */
function startLogTailer(runId: string, logFile: string, admin: ReturnType<typeof createAdminClient>) {
  const FLUSH_INTERVAL_MS = 2_000
  const MAX_WAIT_MS = 4 * 60 * 60 * 1000 // 4 hours max
  let bytesRead = 0
  let logAccum = ""
  let elapsed = 0

  const interval = setInterval(async () => {
    elapsed += FLUSH_INTERVAL_MS
    if (elapsed > MAX_WAIT_MS) {
      clearInterval(interval)
      return
    }

    try {
      const { readFileSync } = await import("fs")
      let fileContent: string
      try {
        fileContent = readFileSync(logFile, "utf8")
      } catch {
        return // file not yet created
      }

      const newContent = fileContent.slice(bytesRead)
      if (!newContent) return
      bytesRead = fileContent.length
      logAccum += newContent

      // Check for exit sentinel
      const exitMatch = logAccum.match(/__EXIT_CODE__:(\d+)/)
      const exitCode = exitMatch ? parseInt(exitMatch[1], 10) : null

      // Strip sentinel from log before writing
      const cleanLog = logAccum.replace(/__EXIT_CODE__:\d+\n?/g, "")

      // Single write — no read-modify-write race condition
      await admin
        .from("local_job_runs")
        .update({ log: cleanLog })
        .eq("id", runId)

      if (exitCode !== null) {
        clearInterval(interval)

        const deliverables = parseDeliverables(cleanLog)
        const status = exitCode === 0 ? "completed" : "failed"

        await admin
          .from("local_job_runs")
          .update({
            status,
            completed_at: new Date().toISOString(),
            deliverables,
          })
          .eq("id", runId)

        // Clean up log file
        try {
          const { unlinkSync } = await import("fs")
          unlinkSync(logFile)
        } catch { /* ignore */ }
      }
    } catch (err) {
      console.error(`[jobs/run] Log tailer error for run ${runId}:`, err)
    }
  }, FLUSH_INTERVAL_MS)
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
