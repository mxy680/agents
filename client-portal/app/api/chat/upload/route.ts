import { NextRequest, NextResponse } from "next/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { verifySession } from "@/lib/session"

/**
 * POST /api/chat/upload
 *
 * Upload a file for use in chat. Stores in Supabase Storage bucket 'chat-files'.
 * Requires a valid session cookie.
 */
export async function POST(request: NextRequest) {
  const cookieValue = request.cookies.get("engagent_session")?.value
  const code = cookieValue ? verifySession(cookieValue) : null
  if (!code) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  const formData = await request.formData()
  const file = formData.get("file") as File | null
  const conversationId = formData.get("conversationId") as string | null

  if (!file) {
    return NextResponse.json({ error: "No file provided" }, { status: 400 })
  }

  // 50MB limit
  if (file.size > 50 * 1024 * 1024) {
    return NextResponse.json({ error: "File too large (50MB max)" }, { status: 400 })
  }

  const admin = createAdminClient()

  // Generate unique path
  const ext = file.name.split(".").pop() ?? "bin"
  const timestamp = Date.now()
  const safeName = file.name.replace(/[^a-zA-Z0-9._-]/g, "_")
  const storagePath = `${conversationId ?? "general"}/${timestamp}_${safeName}`

  // Upload to Supabase Storage
  const buffer = Buffer.from(await file.arrayBuffer())
  const { error: uploadError } = await admin.storage
    .from("chat-files")
    .upload(storagePath, buffer, {
      contentType: file.type || "application/octet-stream",
      upsert: false,
    })

  if (uploadError) {
    console.error("[chat/upload] Upload error:", uploadError.message)
    return NextResponse.json(
      { error: "Upload failed: " + uploadError.message },
      { status: 500 }
    )
  }

  // Get public URL
  const { data: urlData } = admin.storage
    .from("chat-files")
    .getPublicUrl(storagePath)

  return NextResponse.json({
    url: urlData.publicUrl,
    name: file.name,
    size: file.size,
    type: file.type,
    path: storagePath,
  })
}
