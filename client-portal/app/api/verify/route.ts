import { NextRequest, NextResponse } from "next/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { signSession } from "@/lib/session"

const rateLimiter = new Map<string, { count: number; resetAt: number }>()
const RATE_LIMIT = 10
const RATE_WINDOW_MS = 5 * 60 * 1000 // 5 minutes

function checkRateLimit(ip: string): boolean {
  const now = Date.now()
  const entry = rateLimiter.get(ip)

  if (!entry || now > entry.resetAt) {
    rateLimiter.set(ip, { count: 1, resetAt: now + RATE_WINDOW_MS })
    return true
  }

  entry.count++
  return entry.count <= RATE_LIMIT
}

export async function POST(request: NextRequest) {
  const ip =
    request.headers.get("x-forwarded-for") ||
    request.headers.get("x-real-ip") ||
    "unknown"

  if (!checkRateLimit(ip)) {
    return NextResponse.json({ error: "Too many attempts. Try again later." }, { status: 429 })
  }

  let code: string | undefined

  // Try reading code from request body first, then fall back to session cookie
  try {
    const body = await request.json()
    code = body.code || undefined
  } catch {
    // Body parse failure is OK — we may be verifying via cookie
  }

  if (!code) {
    code = request.cookies.get("engagent_session")?.value
  }

  if (!code || typeof code !== "string") {
    return NextResponse.json({ error: "Code is required" }, { status: 400 })
  }

  const admin = createAdminClient()
  const { data, error } = await admin
    .from("client_access")
    .select("id, client_name, agent_name, agent_names, active")
    .eq("code", code.trim())
    .eq("active", true)
    .single()

  if (error || !data) {
    return NextResponse.json({ error: "Invalid access code" }, { status: 401 })
  }

  const agentNames: string[] = (data.agent_names as string[] | null)?.length
    ? (data.agent_names as string[])
    : [data.agent_name]

  const { data: templates } = await admin
    .from("agent_templates")
    .select("name, display_name, description")
    .in("name", agentNames)

  const response = NextResponse.json({
    ok: true,
    clientName: data.client_name,
    agents: (templates ?? []).map((t) => ({
      name: t.name,
      displayName: t.display_name,
      description: t.description,
    })),
  })

  response.cookies.set("engagent_session", signSession(code.trim()), {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "strict",
    maxAge: 60 * 60 * 24 * 30, // 30 days
    path: "/",
  })

  return response
}
