"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"

export default function AuthPage() {
  const router = useRouter()
  const [code, setCode] = useState("")
  const [error, setError] = useState("")
  const [loading, setLoading] = useState(false)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    const trimmed = code.trim()
    if (!trimmed) return

    setLoading(true)
    setError("")

    try {
      const res = await fetch("/api/verify", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ code: trimmed }),
      })

      if (res.ok) {
        // Cookie is set by the verify endpoint — redirect to agents
        router.push("/agents")
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

  return (
    <div className="min-h-screen bg-background flex items-center justify-center p-4">
      <div className="w-full max-w-sm">
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-foreground">Engagent</h1>
          <p className="text-sm text-muted-foreground mt-1">Engineer-as-a-Service</p>
        </div>

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          <div>
            <label
              htmlFor="code"
              className="block text-xs text-muted-foreground mb-1.5"
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
              className="w-full px-3 py-2.5 bg-card border border-border rounded-lg text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:border-ring disabled:opacity-50"
            />
          </div>

          {error && <p className="text-sm text-destructive">{error}</p>}

          <button
            type="submit"
            disabled={loading || !code.trim()}
            className="w-full py-2.5 bg-primary text-primary-foreground rounded-lg text-sm font-medium hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          >
            {loading ? "Verifying..." : "Continue"}
          </button>
        </form>

        <p className="text-xs text-muted-foreground/60 text-center mt-6">
          Don&apos;t have a code? Contact your account manager.
        </p>
      </div>
    </div>
  )
}
