"use client"

import { useEffect, useState } from "react"
import { IconCheck, IconLoader2, IconX } from "@tabler/icons-react"

/**
 * /integrations/canvas/callback
 *
 * The bookmarklet opens this page (same-origin popup) with Canvas cookies
 * in the URL fragment:
 *   /integrations/canvas/callback#base_url=...&cookies=...
 *
 * This page reads the fragment and POSTs to /api/integrations/canvas/save
 * which uses the portal session cookie for auth (same origin = cookies sent).
 */
export default function CanvasCallbackPage() {
  const [status, setStatus] = useState<"saving" | "success" | "error">("saving")
  const [error, setError] = useState("")

  useEffect(() => {
    async function processCookies() {
      const fragment = window.location.hash.slice(1)
      if (!fragment) {
        setError("No data received from bookmarklet")
        setStatus("error")
        return
      }

      const params = new URLSearchParams(fragment)
      const baseUrl = params.get("base_url")
      const cookiesJson = params.get("cookies")

      if (!baseUrl || !cookiesJson) {
        setError("Missing Canvas URL or cookies")
        setStatus("error")
        return
      }

      let cookies: Record<string, string>
      try {
        cookies = JSON.parse(decodeURIComponent(cookiesJson))
      } catch {
        setError("Invalid cookie data")
        setStatus("error")
        return
      }

      // Find session cookie
      const sessionCookie =
        cookies["_normandy_session"] ||
        cookies["canvas_session"] ||
        cookies["_legacy_normandy_session"]

      if (!sessionCookie) {
        setError("No Canvas session cookie found. Make sure you are logged in.")
        setStatus("error")
        return
      }

      // Build credential payload for the save route
      const payload: Record<string, string | undefined> = {
        base_url: decodeURIComponent(baseUrl),
        session_cookie: sessionCookie,
        csrf_token: cookies["_csrf_token"] || undefined,
        log_session_id: cookies["log_session_id"] || undefined,
        label: "Canvas LMS",
      }

      try {
        const res = await fetch("/api/integrations/canvas/save", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        })

        const data = await res.json()
        if (!res.ok) {
          setError(data.error || "Failed to save credentials")
          setStatus("error")
          return
        }

        setStatus("success")
        setTimeout(() => {
          window.location.href = "/integrations"
        }, 2000)
      } catch {
        setError("Network error. Make sure you're logged into the portal.")
        setStatus("error")
      }
    }

    processCookies()
  }, [])

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <div className="flex flex-col items-center gap-4 rounded-lg border bg-card p-8 shadow-sm">
        {status === "saving" && (
          <>
            <IconLoader2 className="size-10 animate-spin text-muted-foreground" />
            <p className="text-lg font-medium">Saving Canvas credentials...</p>
          </>
        )}
        {status === "success" && (
          <>
            <div className="flex size-16 items-center justify-center rounded-full bg-green-100 dark:bg-green-950">
              <IconCheck className="size-8 text-green-600 dark:text-green-400" />
            </div>
            <p className="text-lg font-medium">Canvas LMS connected!</p>
            <p className="text-sm text-muted-foreground">This window will close automatically.</p>
          </>
        )}
        {status === "error" && (
          <>
            <div className="flex size-16 items-center justify-center rounded-full bg-red-100 dark:bg-red-950">
              <IconX className="size-8 text-red-600 dark:text-red-400" />
            </div>
            <p className="text-lg font-medium">Connection failed</p>
            <p className="text-sm text-destructive">{error}</p>
            <p className="text-sm text-muted-foreground">Close this window and try again.</p>
          </>
        )}
      </div>
    </div>
  )
}
