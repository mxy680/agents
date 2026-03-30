export const maxDuration = 30

import { NextRequest, NextResponse } from "next/server"

// Fly.io app URL for remote job triggering
const FLY_APP_URL = "https://engagent.fly.dev"

// Allowlist of agents that can be triggered via this route
const ALLOWED_AGENTS = ["real-estate", "campusreach"]

export async function POST(request: NextRequest) {
  let agent: string
  let job: string
  try {
    const body = await request.json()
    agent = body.agent
    job = body.job
  } catch {
    return NextResponse.json({ error: "Invalid request body" }, { status: 400 })
  }

  if (!agent || !job) {
    return NextResponse.json({ error: "agent and job are required" }, { status: 400 })
  }

  // Allowlist validation — prevent path injection
  if (!ALLOWED_AGENTS.includes(agent)) {
    return NextResponse.json({ error: `Unknown agent: ${agent}` }, { status: 400 })
  }

  try {
    const res = await fetch(`${FLY_APP_URL}/api/jobs/run`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ agent, job }),
    })
    const data = await res.json()
    return NextResponse.json(data, { status: res.status })
  } catch (err) {
    const msg = err instanceof Error ? err.message : String(err)
    return NextResponse.json({ error: `Failed to trigger remote job: ${msg}` }, { status: 502 })
  }
}
