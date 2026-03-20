import { redirect } from "next/navigation"
import { createClient } from "@/lib/supabase/server"

export default async function ChatAgentPage({
  params,
}: {
  params: Promise<{ agentName: string }>
}) {
  const { agentName } = await params

  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) {
    redirect("/login")
  }

  // Create a new conversation and redirect to it
  const { data, error } = await supabase
    .from("conversations")
    .insert({ user_id: user.id, agent_name: agentName })
    .select("id")
    .single()

  if (error || !data) {
    // Fallback: redirect to agents page on failure
    console.error("[chat] Failed to create conversation:", error?.message)
    redirect("/agents")
  }

  redirect(`/chat/${agentName}/${data.id}`)
}
