"use client"

import { useState, useEffect } from "react"
import { useRouter } from "next/navigation"
import { IconRobot, IconArrowRight, IconLoader2 } from "@tabler/icons-react"

interface AgentInfo {
  name: string
  displayName: string
  description: string
}

export default function AgentsPage() {
  const router = useRouter()
  const [clientName, setClientName] = useState("")
  const [agents, setAgents] = useState<AgentInfo[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    async function load() {
      try {
        // Verify session and get agents — cookie sent automatically
        const res = await fetch("/api/verify", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({}),
        })
        if (res.ok) {
          const data = await res.json() as { clientName: string; agents: AgentInfo[] }
          setClientName(data.clientName)
          setAgents(data.agents)
        } else {
          router.push("/auth")
        }
      } catch {
        router.push("/auth")
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [router])

  if (loading) {
    return (
      <div className="min-h-screen bg-background flex items-center justify-center">
        <IconLoader2 className="size-6 text-muted-foreground animate-spin" />
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <div className="w-full max-w-lg">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-foreground">
            Welcome, {clientName}
          </h1>
          <p className="text-sm text-muted-foreground mt-1">
            Select an agent to chat with
          </p>
        </div>

        <div className="flex flex-col gap-3">
          {agents.map((agent) => (
            <button
              key={agent.name}
              onClick={() => router.push(`/chat?agent=${agent.name}`)}
              className="flex items-center gap-4 w-full text-left bg-card border border-border rounded-lg p-4 hover:border-ring transition-colors group"
            >
              <div className="flex size-10 shrink-0 items-center justify-center bg-muted rounded-lg group-hover:bg-accent transition-colors">
                <IconRobot className="size-5 text-muted-foreground" />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-foreground">
                  {agent.displayName}
                </p>
                <p className="text-xs text-muted-foreground mt-0.5 line-clamp-2">
                  {agent.description}
                </p>
              </div>
              <IconArrowRight className="size-4 text-muted-foreground group-hover:text-foreground transition-colors shrink-0" />
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
