import { NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { isAdmin } from "@/lib/admin"

export async function GET() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user || !isAdmin(user.email)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  const admin = createAdminClient()

  // Fetch all active agent templates
  const { data: templates, error } = await admin
    .from("agent_templates")
    .select("id, name, display_name, description, required_integrations, status, created_at")
    .order("display_name", { ascending: true })

  if (error) {
    return NextResponse.json({ error: error.message }, { status: 500 })
  }

  // Fetch active integrations (just provider names)
  const { data: integrations } = await admin
    .from("user_integrations")
    .select("provider")
    .eq("status", "active")

  const connectedProviders = new Set((integrations ?? []).map((i) => i.provider))

  // Fetch client counts per template
  const { data: clientCounts } = await admin
    .from("client_agents")
    .select("template_id")

  const clientCountMap: Record<string, number> = {}
  for (const row of clientCounts ?? []) {
    clientCountMap[row.template_id] = (clientCountMap[row.template_id] ?? 0) + 1
  }

  // Fetch conversation counts per agent
  const { data: convCounts } = await admin
    .from("conversations")
    .select("agent_name")

  const convCountMap: Record<string, number> = {}
  for (const row of convCounts ?? []) {
    convCountMap[row.agent_name] = (convCountMap[row.agent_name] ?? 0) + 1
  }

  // Enrich templates
  const enriched = (templates ?? []).map((t) => {
    const requiredIntegrations = (t.required_integrations ?? []) as string[]
    const integrationHealth = requiredIntegrations.map((provider) => ({
      provider,
      connected: connectedProviders.has(provider),
    }))

    return {
      ...t,
      integrationHealth,
      clientCount: clientCountMap[t.id] ?? 0,
      conversationCount: convCountMap[t.name] ?? 0,
    }
  })

  return NextResponse.json({ agents: enriched })
}
