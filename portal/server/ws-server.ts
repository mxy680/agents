import { WebSocketServer, WebSocket } from "ws"
import { IncomingMessage } from "http"
import { createClient } from "@supabase/supabase-js"
import { createSession, getSession, destroySession, destroyAllSessions } from "../lib/browser-session/manager"
import { encrypt } from "../lib/crypto"
import { verifyOAuthState } from "../lib/oauth-state"
import type { BrowserSession } from "../lib/browser-session/session"
import type { ClientMessage } from "../lib/browser-session/types"

// Validate required env vars at startup
if (!process.env.NEXT_PUBLIC_SUPABASE_URL || !process.env.SUPABASE_SERVICE_ROLE_KEY) {
  console.error("[ws] Missing SUPABASE env vars")
  process.exit(1)
}
if (!process.env.TOKEN_SIGNING_KEY) {
  console.error("[ws] Missing TOKEN_SIGNING_KEY env var")
  process.exit(1)
}
if (!process.env.ENCRYPTION_MASTER_KEY) {
  console.error("[ws] Missing ENCRYPTION_MASTER_KEY env var")
  process.exit(1)
}

// Module-level Supabase client (reused across all connections)
const supabase = createClient(
  process.env.NEXT_PUBLIC_SUPABASE_URL,
  process.env.SUPABASE_SERVICE_ROLE_KEY
)

const PORT = parseInt(process.env.BROWSER_WS_PORT ?? "3001", 10)

/** Validates that a message from the browser client is well-formed. */
function isValidClientMessage(msg: unknown): msg is ClientMessage {
  if (!msg || typeof msg !== "object") return false
  const m = msg as Record<string, unknown>
  switch (m.type) {
    case "click":
    case "mousemove":
    case "mousedown":
    case "mouseup":
      return (
        typeof m.x === "number" &&
        isFinite(m.x) &&
        typeof m.y === "number" &&
        isFinite(m.y)
      )
    case "keydown":
      return typeof m.key === "string" && m.key.length <= 20
    case "keypress":
      return typeof m.text === "string" && m.text.length <= 1
    case "scroll":
      return (
        typeof m.deltaX === "number" &&
        isFinite(m.deltaX) &&
        typeof m.deltaY === "number" &&
        isFinite(m.deltaY)
      )
    case "ping":
      return true
    default:
      return false
  }
}

const wss = new WebSocketServer({ port: PORT })

wss.on("connection", async (ws: WebSocket, req: IncomingMessage) => {
  console.log("[ws] New connection from", req.socket.remoteAddress)
  const rawUrl = req.url ?? "/"
  const url = new URL(rawUrl, `http://localhost`)

  // Path: /browser-session
  if (!url.pathname.startsWith("/browser-session")) {
    ws.close(4000, "Invalid path")
    return
  }

  // Wait for first message to be the auth message (up to 3 seconds)
  let userId: string
  let label: string

  try {
    const authResult = await new Promise<{ userId: string; label: string }>((resolve, reject) => {
      const timeout = setTimeout(() => {
        reject(new Error("Auth timeout"))
      }, 3000)

      ws.once("message", (raw) => {
        clearTimeout(timeout)
        try {
          const msg = JSON.parse(raw.toString())
          if (!msg || msg.type !== "auth" || typeof msg.token !== "string") {
            reject(new Error("Expected auth message"))
            return
          }
          resolve(verifyOAuthState(msg.token))
        } catch (err) {
          reject(err)
        }
      })
    })
    userId = authResult.userId
    label = authResult.label
  } catch (err) {
    console.error("[ws] Auth failed:", err instanceof Error ? err.message : String(err))
    ws.close(4002, "Auth failed")
    return
  }

  // Validate that the userId is a real Supabase user
  const { data: adminData, error: adminError } = await supabase.auth.admin.getUserById(userId)
  if (adminError || !adminData.user) {
    console.error("[ws] User not found:", userId.slice(0, 8) + "...")
    ws.close(4003, "User not found")
    return
  }

  console.log("[ws] Session started for user:", userId.slice(0, 8) + "...")

  let session: BrowserSession
  try {
    session = createSession(userId, label)
    console.log("[ws] Session created:", session.id)
  } catch (err) {
    console.error("[ws] Session creation failed:", err)
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
    console.log("[ws] Starting browser session...")
    await session.start()
    console.log("[ws] Browser session started successfully")
  } catch (err) {
    console.error("[ws] Session start failed:", err)
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
      if (!isValidClientMessage(msg)) return
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

process.on("SIGTERM", () => {
  destroyAllSessions()
  process.exit(0)
})

process.on("SIGINT", () => {
  destroyAllSessions()
  process.exit(0)
})

console.log(`Browser session WS server running on port ${PORT}`)
