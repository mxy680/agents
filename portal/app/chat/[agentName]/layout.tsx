import { createClient } from "@/lib/supabase/server"
import { ChatSidebar } from "@/components/chat-sidebar"
import { redirect } from "next/navigation"

interface Conversation {
  id: string
  user_id: string
  agent_name: string
  title: string | null
  starred: boolean
  session_id: string | null
  created_at: string
  updated_at: string
}

export default async function ChatAgentLayout({
  children,
  params,
}: {
  children: React.ReactNode
  params: Promise<{ agentName: string }>
}) {
  const { agentName } = await params

  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) {
    redirect("/login")
  }

  const { data } = await supabase
    .from("conversations")
    .select("*")
    .eq("user_id", user.id)
    .eq("agent_name", agentName)
    .order("updated_at", { ascending: false })

  const conversations = (data ?? []) as Conversation[]

  return (
    <div className="flex h-screen overflow-hidden">
      <ChatSidebar agentName={agentName} conversations={conversations} />
      <div className="flex-1 overflow-hidden">{children}</div>
    </div>
  )
}
