"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"

export default function ClientEntryPage() {
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
      const res = await fetch(`/api/client/verify`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ code: trimmed }),
      })

      if (res.ok) {
        router.push(`/client/chat/${encodeURIComponent(trimmed)}`)
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

          {error && (
            <p className="text-sm text-red-400">{error}</p>
          )}

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
