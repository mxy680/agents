"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { IconRobot, IconArrowRight } from "@tabler/icons-react"

interface AgentInfo {
  name: string
  displayName: string
  description: string
}

export default function ClientEntryPage() {
  const router = useRouter()
  const [code, setCode] = useState("")
  const [error, setError] = useState("")
  const [loading, setLoading] = useState(false)
  const [clientName, setClientName] = useState("")
  const [agents, setAgents] = useState<AgentInfo[]>([])
  const [verified, setVerified] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const trimmed = code.trim()
    if (!trimmed) return

    setLoading(true)
    setError("")

    try {
      const res = await fetch("/api/client/verify", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ code: trimmed }),
      })

      if (res.ok) {
        const data = await res.json() as {
          clientName: string
          agents: AgentInfo[]
        }
        setClientName(data.clientName)
        setAgents(data.agents)

        if (data.agents.length === 1) {
          // Single agent — go straight to chat
          router.push(`/client/chat/${encodeURIComponent(trimmed)}?agent=${data.agents[0].name}`)
        } else {
          // Multiple agents — show selection
          setVerified(true)
        }
      } else {
        const data = await res.json().catch(() => ({})) as { error?: string }
        setError(data.error ?? "Invalid code")
      }
    } catch {
      setError("Connection error. Please try again.")
    } finally {
      setLoading(false)
    }
  }

  if (verified) {
    return (
      <div className="min-h-screen bg-[#0a0a0a] flex items-center justify-center p-4">
        <div className="w-full max-w-lg">
          <div className="text-center mb-8">
            <h1 className="text-2xl font-bold text-white">
              Welcome, {clientName}
            </h1>
            <p className="text-sm text-neutral-500 mt-1">
              Select an agent to chat with
            </p>
          </div>

          <div className="flex flex-col gap-3">
            {agents.map((agent) => (
              <button
                key={agent.name}
                onClick={() =>
                  router.push(
                    `/client/chat/${encodeURIComponent(code.trim())}?agent=${agent.name}`
                  )
                }
                className="flex items-center gap-4 w-full text-left bg-neutral-900 border border-neutral-800 rounded-lg p-4 hover:border-neutral-600 transition-colors group"
              >
                <div className="flex size-10 shrink-0 items-center justify-center bg-neutral-800 rounded-lg group-hover:bg-neutral-700 transition-colors">
                  <IconRobot className="size-5 text-neutral-400" />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-white">
                    {agent.displayName}
                  </p>
                  <p className="text-xs text-neutral-500 mt-0.5 line-clamp-2">
                    {agent.description}
                  </p>
                </div>
                <IconArrowRight className="size-4 text-neutral-600 group-hover:text-neutral-400 transition-colors shrink-0" />
              </button>
            ))}
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-[#0a0a0a] flex items-center justify-center p-4">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-white">Emdash</h1>
          <p className="text-sm text-neutral-500 mt-1">AI Agent Portal</p>
        </div>

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div>
            <label
              htmlFor="code"
              className="block text-xs text-neutral-400 mb-1.5"
            >
              Access Code
            </label>
            <input
              id="code"
              type="text"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              placeholder="Enter your access code"
              autoFocus
              disabled={loading}
              className="w-full px-3 py-2.5 bg-neutral-900 border border-neutral-800 rounded-lg text-sm text-white placeholder:text-neutral-600 focus:outline-none focus:border-neutral-600 disabled:opacity-50"
            />
          </div>

          {error && <p className="text-sm text-red-400">{error}</p>}

          <button
            type="submit"
            disabled={loading || !code.trim()}
            className="w-full py-2.5 bg-white text-black rounded-lg text-sm font-medium hover:bg-neutral-200 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {loading ? "Verifying..." : "Continue"}
          </button>
        </form>

        <p className="text-xs text-neutral-600 text-center mt-6">
          Don&apos;t have a code? Contact your account manager.
        </p>
      </div>
    </div>
  )
}
