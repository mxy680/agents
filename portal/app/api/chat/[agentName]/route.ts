import { NextRequest } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { resolveUserCredentials } from "@/lib/credentials"
import { runContainer } from "@/lib/container-runner"
import path from "path"
import os from "os"

export async function POST(
  request: NextRequest,
  { params }: { params: Promise<{ agentName: string }> }
) {
  const { agentName } = await params

  // Validate agentName: only allow alphanumeric, hyphens, underscores
  if (!/^[a-z0-9-_]+$/i.test(agentName)) {
    return new Response(JSON.stringify({ error: "Invalid agent name" }), { status: 400 })
  }

  // Authenticate user
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()
  if (!user) {
    return new Response(JSON.stringify({ error: "Unauthorized" }), { status: 401 })
  }

  // Check that the user has acquired this agent
  const { data: template } = await supabase
    .from("agent_templates")
    .select("id")
    .eq("name", agentName)
    .eq("status", "active")
    .single()

  if (template) {
    const { data: userAgent } = await supabase
      .from("user_agents")
      .select("template_id, status")
      .eq("user_id", user.id)
      .eq("template_id", template.id)
      .maybeSingle()

    if (!userAgent) {
      return new Response(
        JSON.stringify({ error: "Agent not acquired. Visit the marketplace to get this agent." }),
        { status: 403 }
      )
    }

    if (userAgent.status === "pending") {
      return new Response(
        JSON.stringify({ error: "Your access to this agent is pending admin approval." }),
        { status: 403 }
      )
    }

    if (userAgent.status === "rejected") {
      return new Response(
        JSON.stringify({ error: "Your access to this agent was not approved." }),
        { status: 403 }
      )
    }
  }

  // Parse request body
  let message: string
  let sessionId: string | undefined
  try {
    const body = await request.json()
    message = body.message
    sessionId = body.sessionId
  } catch {
    return new Response(JSON.stringify({ error: "Invalid request body" }), { status: 400 })
  }

  if (!message || typeof message !== "string") {
    return new Response(JSON.stringify({ error: "message is required" }), { status: 400 })
  }

  // Resolve credentials for all active integrations
  let credEnv: Record<string, string>
  try {
    credEnv = await resolveUserCredentials(user.id)
  } catch (e) {
    console.error("Failed to resolve user credentials:", e)
    return new Response(JSON.stringify({ error: "Failed to fetch integrations" }), { status: 500 })
  }

  // CLAUDE_CODE_OAUTH_TOKEN is required
  const claudeToken = process.env.CLAUDE_CODE_OAUTH_TOKEN
  if (!claudeToken) {
    return new Response(JSON.stringify({ error: "CLAUDE_CODE_OAUTH_TOKEN not configured" }), { status: 500 })
  }

  // Build instance directory
  const instancePath = path.join(
    os.tmpdir(),
    "agents",
    user.id,
    agentName,
    String(Date.now())
  )

  // Copy agent template files if they exist
  try {
    const { mkdirSync, copyFileSync, existsSync } = await import("fs")
    mkdirSync(instancePath, { recursive: true })

    const agentsDir = path.join(process.cwd(), "..", "agents", agentName)
    for (const file of ["role.md", "CLAUDE.md"]) {
      const src = path.join(agentsDir, file)
      if (existsSync(src)) {
        copyFileSync(src, path.join(instancePath, file))
      }
    }
  } catch (e) {
    console.error("Failed to copy agent template files:", e)
  }

  const containerEnv: Record<string, string> = {
    CLAUDE_CODE_OAUTH_TOKEN: claudeToken,
    ...credEnv,
  }

  // Set up SSE stream
  const stream = new ReadableStream({
    async start(controller) {
      const encoder = new TextEncoder()

      function send(event: string, data: string) {
        // SSE spec: multi-line data must be sent as multiple "data:" lines
        const lines = data.split("\n").map((l) => `data: ${l}`).join("\n")
        controller.enqueue(encoder.encode(`event: ${event}\n${lines}\n\n`))
      }

      const abortController = new AbortController()

      // If client disconnects, abort the container
      request.signal.addEventListener("abort", () => {
        abortController.abort()
      })

      try {
        for await (const sseEvent of runContainer({
          instancePath,
          message,
          sessionId,
          env: containerEnv,
          signal: abortController.signal,
          timeoutMs: 5 * 60 * 1000, // 5 minute timeout
        })) {
          send(sseEvent.event, sseEvent.data)
        }
      } catch (e) {
        const msg = e instanceof Error ? e.message : String(e)
        send("error", msg)
      } finally {
        controller.close()
      }
    },
  })

  return new Response(stream, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache, no-transform",
      Connection: "keep-alive",
      "X-Accel-Buffering": "no",
    },
  })
}
