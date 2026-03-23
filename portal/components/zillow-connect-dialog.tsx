"use client"

import { useState } from "react"
import { useRouter } from "next/navigation"
import { Button } from "@/components/ui/button"
import { IconLoader2, IconCheck, IconAlertCircle } from "@tabler/icons-react"

interface ZillowConnectDialogProps {
  children: React.ReactNode
}

export function ZillowConnectDialog({ children }: ZillowConnectDialogProps) {
  const router = useRouter()
  const [status, setStatus] = useState<"idle" | "loading" | "done" | "error">("idle")
  const [message, setMessage] = useState("")

  async function handleCapture() {
    setStatus("loading")
    setMessage("Browser opening — solve the CAPTCHA if prompted...")

    try {
      const controller = new AbortController()
      const timeout = setTimeout(() => controller.abort(), 180_000) // 3 min
      const res = await fetch("/api/integrations/zillow/refresh", {
        method: "POST",
        signal: controller.signal,
      })
      clearTimeout(timeout)
      const data = await res.json()

      if (data.ok) {
        setStatus("done")
        setMessage(`Captured ${data.cookieCount} cookies`)
        setTimeout(() => {
          setStatus("idle")
          setMessage("")
          router.refresh()
        }, 1500)
      } else {
        setStatus("error")
        setMessage(data.error || "Failed to capture cookies")
      }
    } catch {
      setStatus("error")
      setMessage("Network error")
    }
  }

  if (status === "idle") {
    return <div onClick={handleCapture}>{children}</div>
  }

  return (
    <div className="flex items-center justify-center gap-2 py-1 text-sm w-full">
      {status === "loading" && (
        <>
          <IconLoader2 className="size-4 animate-spin text-muted-foreground" />
          <span className="text-muted-foreground">{message}</span>
        </>
      )}
      {status === "done" && (
        <>
          <IconCheck className="size-4 text-green-500" />
          <span className="text-green-500">{message}</span>
        </>
      )}
      {status === "error" && (
        <div className="flex flex-col items-center gap-1">
          <div className="flex items-center gap-2">
            <IconAlertCircle className="size-4 text-red-500" />
            <span className="text-red-500">{message}</span>
          </div>
          <Button variant="ghost" size="sm" onClick={() => { setStatus("idle"); setMessage("") }}>
            Try again
          </Button>
        </div>
      )}
    </div>
  )
}
