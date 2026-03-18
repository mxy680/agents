import { NextRequest } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"

export async function POST(request: NextRequest) {
  // Authenticate user
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) {
    return Response.json({ error: "Unauthorized" }, { status: 401 })
  }

  // Parse request body
  let templateId: string
  try {
    const body = await request.json()
    templateId = body.templateId
  } catch {
    return Response.json({ error: "Invalid request body" }, { status: 400 })
  }

  if (!templateId || typeof templateId !== "string") {
    return Response.json({ error: "templateId is required" }, { status: 400 })
  }

  // Validate template exists and is active
  const { data: template, error: templateError } = await supabase
    .from("agent_templates")
    .select("id, status")
    .eq("id", templateId)
    .single()

  if (templateError || !template) {
    return Response.json({ error: "Template not found" }, { status: 404 })
  }

  if (template.status !== "active") {
    return Response.json({ error: "Template is not available" }, { status: 400 })
  }

  // Upsert into user_agents (handles duplicate gracefully)
  const { error: upsertError } = await supabase
    .from("user_agents")
    .upsert(
      { user_id: user.id, template_id: templateId },
      { onConflict: "user_id,template_id", ignoreDuplicates: true }
    )

  if (upsertError) {
    console.error("Failed to acquire agent:", upsertError)
    return Response.json({ error: "Failed to acquire agent" }, { status: 500 })
  }

  // Increment acquisition_count on the template (non-fatal if it fails)
  // Use admin client to bypass RLS on agent_templates (read-only for users)
  const admin = createAdminClient()
  const { data: current } = await admin
    .from("agent_templates")
    .select("acquisition_count")
    .eq("id", templateId)
    .single()

  if (current) {
    await admin
      .from("agent_templates")
      .update({ acquisition_count: (current.acquisition_count ?? 0) + 1 })
      .eq("id", templateId)
  }

  return Response.json({ acquired: true })
}
