import { WebSocketServer, WebSocket } from "ws"
import { IncomingMessage } from "http"
import { createClient } from "@supabase/supabase-js"
import { createSession, getSession, destroySession } from "../lib/browser-session/manager"
import { encrypt } from "../lib/crypto"
import { verifyOAuthState } from "../lib/oauth-state"
import type { BrowserSession } from "../lib/browser-session/session"

const PORT = parseInt(process.env.BROWSER_WS_PORT ?? "3001", 10)

const wss = new WebSocketServer({ port: PORT })

wss.on("connection", async (ws: WebSocket, req: IncomingMessage) => {
  const rawUrl = req.url ?? "/"
  const url = new URL(rawUrl, `http://localhost`)

  // Path: /browser-session  (session created on connect using token)
  if (!url.pathname.startsWith("/browser-session")) {
    ws.close(4000, "Invalid path")
    return
  }

  const token = url.searchParams.get("token")

  if (!token) {
    ws.close(4002, "Missing auth token")
    return
  }

  let userId: string
  let label: string
  try {
    const verified = verifyOAuthState(token)
    userId = verified.userId
    label = verified.label
  } catch (err) {
    ws.close(4003, `Invalid token: ${String(err)}`)
    return
  }

  let session: BrowserSession
  try {
    session = createSession(userId, label)
  } catch (err) {
    ws.close(4004, String(err))
    return
  }

  function send(data: object): void {
    if (ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(data))
    }
  }

  session.setHandlers({
    onFrame: (data) => send({ type: "frame", data }),
    onStatus: (status) => send({ type: "status", status }),
    onCookies: async (cookies) => {
      try {
        const payload = JSON.stringify(cookies)
        const encrypted = encrypt(payload)

        const supabase = createClient(
          process.env.NEXT_PUBLIC_SUPABASE_URL!,
          process.env.SUPABASE_SERVICE_ROLE_KEY!
        )

        await supabase.from("user_integrations").upsert(
          {
            user_id: userId,
            provider: "instagram",
            label,
            status: "active",
            credentials: `\\x${encrypted.toString("hex")}`,
            updated_at: new Date().toISOString(),
          },
          { onConflict: "user_id,provider,label" }
        )

        send({ type: "cookies", success: true })
        send({ type: "status", status: "complete" })
      } catch (err) {
        send({ type: "cookies", success: false, error: String(err) })
      }

      setTimeout(() => destroySession(session.id), 2000)
    },
  })

  try {
    await session.start()
  } catch (err) {
    send({ type: "status", status: "error" })
    ws.close(4005, `Session start failed: ${String(err)}`)
    destroySession(session.id)
    return
  }

  send({ type: "viewport", width: 1280, height: 720 })

  ws.on("message", (raw) => {
    try {
      const msg = JSON.parse(raw.toString())
      if (msg.type === "ping") return
      session.handleInput(msg).catch(() => {})
    } catch {
      // Ignore malformed messages
    }
  })

  ws.on("close", () => {
    // Grace period before destroying so brief reconnects work
    setTimeout(() => {
      if (getSession(session.id)) {
        destroySession(session.id)
      }
    }, 10_000)
  })
})

console.log(`Browser session WS server running on port ${PORT}`)
