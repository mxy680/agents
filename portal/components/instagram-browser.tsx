"use client"

import { useEffect, useRef, useState, useCallback } from "react"
import { Button } from "@/components/ui/button"
import type { ServerMessage, ClientMessage, SessionStatus } from "@/lib/browser-session/types"

interface InstagramBrowserProps {
  wsUrl: string
  onComplete: () => void
  onCancel: () => void
}

const STATUS_MESSAGES: Record<SessionStatus, string> = {
  loading: "Starting secure browser...",
  ready: "Browser ready — log in to Instagram",
  login_detected: "Login detected...",
  extracting: "Extracting session...",
  complete: "Connected! Closing...",
  error: "An error occurred. Please try again.",
  timeout: "Session timed out. Please try again.",
}

export function InstagramBrowser({ wsUrl, onComplete, onCancel }: InstagramBrowserProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const lastMoveRef = useRef<number>(0)
  const [viewport, setViewport] = useState({ width: 1280, height: 720 })
  const [status, setStatus] = useState<SessionStatus>("loading")
  const [isDone, setIsDone] = useState(false)

  const sendMsg = useCallback((msg: ClientMessage) => {
    const ws = wsRef.current
    if (ws && ws.readyState === WebSocket.OPEN) {
      ws.send(JSON.stringify(msg))
    }
  }, [])

  useEffect(() => {
    const ws = new WebSocket(wsUrl)
    wsRef.current = ws

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data) as ServerMessage
        switch (msg.type) {
          case "viewport":
            setViewport({ width: msg.width, height: msg.height })
            break

          case "frame": {
            const canvas = canvasRef.current
            if (!canvas) break
            const ctx = canvas.getContext("2d")
            if (!ctx) break
            const img = new Image()
            img.onload = () => {
              ctx.drawImage(img, 0, 0, canvas.width, canvas.height)
            }
            img.src = `data:image/jpeg;base64,${msg.data}`
            break
          }

          case "status":
            setStatus(msg.status)
            if (msg.status === "complete") {
              setIsDone(true)
              setTimeout(() => onComplete(), 1500)
            }
            break

          case "cookies":
            // Handled via status message; nothing extra needed client-side
            break
        }
      } catch {
        // Ignore parse errors
      }
    }

    // Keep-alive ping every 20 seconds
    const pingInterval = setInterval(() => {
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ type: "ping" }))
      }
    }, 20_000)

    return () => {
      clearInterval(pingInterval)
      ws.close()
      wsRef.current = null
    }
  }, [wsUrl, onComplete])

  function getScaledCoords(
    e: React.MouseEvent<HTMLCanvasElement>
  ): { x: number; y: number } {
    const canvas = canvasRef.current!
    const rect = canvas.getBoundingClientRect()
    const scaleX = viewport.width / rect.width
    const scaleY = viewport.height / rect.height
    return {
      x: Math.round((e.clientX - rect.left) * scaleX),
      y: Math.round((e.clientY - rect.top) * scaleY),
    }
  }

  function handleClick(e: React.MouseEvent<HTMLCanvasElement>) {
    const { x, y } = getScaledCoords(e)
    sendMsg({ type: "click", x, y })
  }

  function handleMouseMove(e: React.MouseEvent<HTMLCanvasElement>) {
    const now = Date.now()
    if (now - lastMoveRef.current < 50) return
    lastMoveRef.current = now
    const { x, y } = getScaledCoords(e)
    sendMsg({ type: "mousemove", x, y })
  }

  function handleMouseDown(e: React.MouseEvent<HTMLCanvasElement>) {
    const { x, y } = getScaledCoords(e)
    sendMsg({ type: "mousedown", x, y })
  }

  function handleMouseUp(e: React.MouseEvent<HTMLCanvasElement>) {
    const { x, y } = getScaledCoords(e)
    sendMsg({ type: "mouseup", x, y })
  }

  function handleWheel(e: React.WheelEvent<HTMLCanvasElement>) {
    e.preventDefault()
    sendMsg({ type: "scroll", deltaX: e.deltaX, deltaY: e.deltaY })
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLCanvasElement>) {
    e.preventDefault()
    if (e.key.length === 1 && !e.ctrlKey && !e.metaKey && !e.altKey) {
      // Printable character — use keypress only
      sendMsg({ type: "keypress", text: e.key })
    } else {
      // Special keys (Enter, Backspace, Tab, arrows, etc.)
      sendMsg({ type: "keydown", key: e.key })
    }
  }

  function handleCancel() {
    wsRef.current?.close()
    onCancel()
  }

  const isError = status === "error" || status === "timeout"

  return (
    <div className="flex min-h-0 flex-1 flex-col gap-3">
      {/* Status bar */}
      <div className="flex items-center gap-2 rounded-md bg-muted px-3 py-2 text-sm">
        {!isError && !isDone && (
          <span className="inline-block size-2 animate-pulse rounded-full bg-green-500" />
        )}
        {isError && <span className="inline-block size-2 rounded-full bg-red-500" />}
        {isDone && <span className="inline-block size-2 rounded-full bg-green-500" />}
        <span className={isError ? "text-destructive" : "text-muted-foreground"}>
          {STATUS_MESSAGES[status]}
        </span>
      </div>

      {/* Browser canvas */}
      <div className="relative min-h-0 flex-1 overflow-hidden rounded-md border bg-black">
        <canvas
          ref={canvasRef}
          width={viewport.width}
          height={viewport.height}
          className="size-full cursor-pointer object-contain"
          tabIndex={0}
          onClick={handleClick}
          onMouseMove={handleMouseMove}
          onMouseDown={handleMouseDown}
          onMouseUp={handleMouseUp}
          onWheel={handleWheel}
          onKeyDown={handleKeyDown}
        />
        {status === "loading" && (
          <div className="absolute inset-0 flex items-center justify-center bg-black/60">
            <p className="text-sm text-white">Starting browser...</p>
          </div>
        )}
      </div>

      {/* Footer */}
      <div className="flex items-center justify-between">
        <p className="text-xs text-muted-foreground">
          Secure session — your password is never stored
        </p>
        <Button variant="outline" size="sm" onClick={handleCancel} disabled={isDone}>
          Cancel
        </Button>
      </div>
    </div>
  )
}
