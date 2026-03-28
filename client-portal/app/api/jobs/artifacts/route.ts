import { NextRequest, NextResponse } from "next/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { readFileSync } from "fs"
import path from "path"

const MIME_MAP: Record<string, string> = {
  ".pdf": "application/pdf",
  ".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
  ".xls": "application/vnd.ms-excel",
  ".csv": "text/csv",
  ".json": "application/json",
  ".txt": "text/plain",
  ".png": "image/png",
  ".jpg": "image/jpeg",
  ".jpeg": "image/jpeg",
}

function lookupMime(filename: string): string {
  const ext = path.extname(filename).toLowerCase()
  return MIME_MAP[ext] ?? "application/octet-stream"
}

/**
 * POST /api/jobs/artifacts
 *
 * Upload a job artifact (PDF, XLSX, etc.) to Supabase Storage.
 * Called by job scripts/agents with either:
 *   - multipart form: file + runId
 *   - JSON body: { runId, filePath } (for server-side files)
 *
 * Returns the public URL and updates the job run's deliverables.
 */
export async function POST(request: NextRequest) {
  const admin = createAdminClient()

  let fileBuffer: Buffer
  let fileName: string
  let contentType: string
  let runId: string

  const ct = request.headers.get("content-type") ?? ""

  if (ct.includes("multipart/form-data")) {
    const formData = await request.formData()
    const file = formData.get("file") as File | null
    runId = (formData.get("runId") as string) ?? ""

    if (!file || !runId) {
      return NextResponse.json({ error: "file and runId required" }, { status: 400 })
    }

    fileBuffer = Buffer.from(await file.arrayBuffer())
    fileName = file.name
    contentType = file.type || "application/octet-stream"
  } else {
    // JSON body with filePath (agent uploads from local filesystem)
    const body = await request.json()
    runId = body.runId
    const filePath: string = body.filePath

    if (!runId || !filePath) {
      return NextResponse.json({ error: "runId and filePath required" }, { status: 400 })
    }

    try {
      fileBuffer = readFileSync(filePath)
    } catch {
      return NextResponse.json({ error: `File not found: ${filePath}` }, { status: 404 })
    }

    fileName = path.basename(filePath)
    contentType = lookupMime(fileName)
  }

  // 100MB limit
  if (fileBuffer.length > 100 * 1024 * 1024) {
    return NextResponse.json({ error: "File too large (100MB max)" }, { status: 400 })
  }

  // Upload to Supabase Storage
  const safeName = fileName.replace(/[^a-zA-Z0-9._-]/g, "_")
  const storagePath = `${runId}/${Date.now()}_${safeName}`

  const { error: uploadError } = await admin.storage
    .from("job-artifacts")
    .upload(storagePath, fileBuffer, {
      contentType,
      upsert: false,
    })

  if (uploadError) {
    console.error("[jobs/artifacts] Upload error:", uploadError.message)
    return NextResponse.json({ error: "Upload failed: " + uploadError.message }, { status: 500 })
  }

  // Get public URL
  const { data: urlData } = admin.storage
    .from("job-artifacts")
    .getPublicUrl(storagePath)

  const publicUrl = urlData.publicUrl

  // Update job run deliverables
  const { data: run } = await admin
    .from("local_job_runs")
    .select("deliverables")
    .eq("id", runId)
    .single()

  const existing = (run?.deliverables as Record<string, string>) ?? {}
  const ext = path.extname(fileName).toLowerCase()
  const key = ext === ".pdf" ? "pdf_url" : ext === ".xlsx" ? "sheet_url" : `artifact_${safeName}`

  await admin
    .from("local_job_runs")
    .update({ deliverables: { ...existing, [key]: publicUrl } })
    .eq("id", runId)

  return NextResponse.json({
    url: publicUrl,
    name: fileName,
    size: fileBuffer.length,
    type: contentType,
    path: storagePath,
    key,
  })
}
