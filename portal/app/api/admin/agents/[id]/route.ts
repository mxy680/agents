import { NextRequest } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { isAdmin } from "@/lib/admin"

export async function PATCH(
  request: NextRequest,
  { params }: { params: Promise<{ id: string }> }
) {
  const { id } = await params

  // Authenticate caller
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) {
    return Response.json({ error: "Unauthorized" }, { status: 401 })
  }

  if (!isAdmin(user.email)) {
    return Response.json({ error: "Forbidden" }, { status: 403 })
  }

  // Parse body
  let action: string
  let note: string | undefined
  try {
    const body = await request.json()
    action = body.action
    note = body.note
  } catch {
    return Response.json({ error: "Invalid request body" }, { status: 400 })
  }

  if (action !== "approve" && action !== "reject") {
    return Response.json({ error: "action must be 'approve' or 'reject'" }, { status: 400 })
  }

  const newStatus = action === "approve" ? "approved" : "rejected"

  const admin = createAdminClient()
  const { error } = await admin
    .from("user_agents")
    .update({
      status: newStatus,
      reviewed_at: new Date().toISOString(),
      reviewer_note: note ?? null,
    })
    .eq("id", id)

  if (error) {
    console.error("Failed to update user_agent:", error)
    return Response.json({ error: "Failed to update request" }, { status: 500 })
  }

  return Response.json({ success: true, status: newStatus })
}
