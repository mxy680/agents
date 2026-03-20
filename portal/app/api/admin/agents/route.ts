import { NextRequest } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { isAdmin } from "@/lib/admin"

export async function GET(request: NextRequest) {
  // Authenticate caller
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) {
    return Response.json({ error: "Unauthorized" }, { status: 401 })
  }

  if (!isAdmin(user.email)) {
    return Response.json({ error: "Forbidden" }, { status: 403 })
  }

  const { searchParams } = new URL(request.url)
  const statusFilter = searchParams.get("status") ?? "pending"

  const admin = createAdminClient()

  // Build query with optional status filter
  let query = admin
    .from("user_agents")
    .select("id, user_id, template_id, status, acquired_at, reviewed_at, reviewer_note, agent_templates(id, name, display_name)")
    .order("acquired_at", { ascending: true })

  if (statusFilter !== "all") {
    query = query.eq("status", statusFilter)
  }

  const { data: rows, error } = await query

  if (error) {
    console.error("Failed to fetch user_agents:", error)
    return Response.json({ error: "Failed to fetch requests" }, { status: 500 })
  }

  // Fetch emails for each unique user_id via admin auth API
  const userIds = Array.from(new Set((rows ?? []).map((r) => r.user_id)))
  const emailMap: Record<string, string> = {}

  for (const uid of userIds) {
    try {
      const { data: userData } = await admin.auth.admin.getUserById(uid)
      if (userData.user?.email) {
        emailMap[uid] = userData.user.email
      }
    } catch {
      // non-fatal: email stays unknown
    }
  }

  const result = (rows ?? []).map((row) => {
    const tmpl = Array.isArray(row.agent_templates)
      ? row.agent_templates[0]
      : row.agent_templates
    return {
      id: row.id,
      user_id: row.user_id,
      user_email: emailMap[row.user_id] ?? "Unknown",
      template_id: row.template_id,
      template_name: tmpl?.name ?? "",
      template_display_name: tmpl?.display_name ?? "",
      status: row.status,
      acquired_at: row.acquired_at,
      reviewed_at: row.reviewed_at,
      reviewer_note: row.reviewer_note,
    }
  })

  return Response.json({ requests: result })
}
