"use client"

import * as React from "react"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { Field, FieldGroup, FieldLabel } from "@/components/ui/field"
import { IconCheck, IconAlertCircle, IconLoader2, IconClipboard } from "@tabler/icons-react"

interface CookiePasteDialogProps {
  provider: string
  providerName: string
  children: React.ReactNode
}

type Status = "idle" | "saving" | "done" | "error"

export function CookiePasteDialog({
  provider,
  providerName,
  children,
}: CookiePasteDialogProps) {
  const [open, setOpen] = React.useState(false)
  const [label, setLabel] = React.useState("")
  const [cookieText, setCookieText] = React.useState("")
  const [baseUrl, setBaseUrl] = React.useState("")
  const [status, setStatus] = React.useState<Status>("idle")
  const [message, setMessage] = React.useState("")
  const [parseError, setParseError] = React.useState("")

  const needsBaseUrl = provider === "canvas"

  function handleOpenChange(next: boolean) {
    if (!next) {
      setStatus("idle")
      setMessage("")
      setLabel("")
      setCookieText("")
      setBaseUrl("")
      setParseError("")
    }
    setOpen(next)
  }

  function validateJson(text: string): Record<string, string> | null {
    if (!text.trim()) {
      setParseError("Paste your cookies JSON here")
      return null
    }
    try {
      const parsed = JSON.parse(text)
      if (typeof parsed !== "object" || Array.isArray(parsed) || parsed === null) {
        setParseError("Expected a JSON object (e.g. { \"cookie_name\": \"value\" })")
        return null
      }
      setParseError("")
      return parsed as Record<string, string>
    } catch {
      setParseError("Invalid JSON — use the Emdash Cookie Copier extension to copy cookies")
      return null
    }
  }

  function handleCookieTextChange(text: string) {
    setCookieText(text)
    if (text.trim()) {
      validateJson(text)
    } else {
      setParseError("")
    }
  }

  async function handleSubmit() {
    const cookies = validateJson(cookieText)
    if (!cookies) return

    const cookieCount = Object.keys(cookies).length
    if (cookieCount === 0) {
      setParseError("The cookies object is empty")
      return
    }

    if (needsBaseUrl && !baseUrl.trim()) {
      setParseError("Canvas base URL is required")
      return
    }

    const accountLabel = label.trim() || `${providerName} Account`
    setStatus("saving")
    setMessage("Saving credentials...")

    try {
      const res = await fetch("/api/integrations/save-cookies", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          provider,
          label: accountLabel,
          cookies,
          ...(needsBaseUrl && baseUrl.trim() ? { baseUrl: baseUrl.trim() } : {}),
        }),
      })

      const data = await res.json()

      if (!res.ok) {
        setStatus("error")
        setMessage(data.error || "Failed to save cookies")
        return
      }

      setStatus("done")
      setMessage(`${accountLabel} connected (${cookieCount} cookies)`)
      setTimeout(() => {
        setOpen(false)
        window.location.reload()
      }, 1500)
    } catch {
      setStatus("error")
      setMessage("Network error — please try again")
    }
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Connect {providerName}</DialogTitle>
          <DialogDescription>
            Use the Emdash Cookie Copier extension on {providerName}&apos;s website, then paste the
            copied JSON below.
          </DialogDescription>
        </DialogHeader>

        {status === "idle" && (
          <>
            <div className="flex flex-col gap-4 py-2">
              <FieldGroup>
                {needsBaseUrl && (
                  <Field>
                    <FieldLabel htmlFor="cookie-base-url">Canvas URL</FieldLabel>
                    <div className="flex gap-2">
                      <Input
                        id="cookie-base-url"
                        placeholder="https://canvas.university.edu"
                        value={baseUrl}
                        onChange={(e) => setBaseUrl(e.target.value)}
                        autoFocus
                      />
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        className="shrink-0"
                        onClick={() => setBaseUrl("https://canvas.case.edu")}
                      >
                        CWRU
                      </Button>
                    </div>
                  </Field>
                )}
                <Field>
                  <FieldLabel htmlFor="cookie-label">Account name</FieldLabel>
                  <Input
                    id="cookie-label"
                    placeholder={`e.g. Personal ${providerName}`}
                    value={label}
                    onChange={(e) => setLabel(e.target.value)}
                  />
                </Field>
                <Field>
                  <FieldLabel htmlFor="cookie-json">
                    Cookies JSON
                  </FieldLabel>
                  <Textarea
                    id="cookie-json"
                    placeholder='{ "cookie_name": "value", ... }'
                    value={cookieText}
                    onChange={(e) => handleCookieTextChange(e.target.value)}
                    className="font-mono text-xs min-h-[120px]"
                    autoFocus
                  />
                  {parseError && (
                    <p className="text-xs text-destructive mt-1">{parseError}</p>
                  )}
                </Field>
              </FieldGroup>
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button
                onClick={handleSubmit}
                disabled={!cookieText.trim() || !!parseError}
              >
                <IconClipboard className="mr-2 size-4" />
                Save Cookies
              </Button>
            </DialogFooter>
          </>
        )}

        {status === "saving" && (
          <div className="flex items-center justify-center gap-2 py-4 text-sm text-muted-foreground">
            <IconLoader2 className="size-4 animate-spin" />
            {message}
          </div>
        )}

        {status === "done" && (
          <div className="flex items-center justify-center gap-2 py-4 text-sm text-green-500">
            <IconCheck className="size-4" />
            {message}
          </div>
        )}

        {status === "error" && (
          <div className="flex flex-col items-center gap-3 py-4">
            <div className="flex items-center gap-2 text-sm text-destructive">
              <IconAlertCircle className="size-4" />
              {message}
            </div>
            <Button variant="outline" size="sm" onClick={() => setStatus("idle")}>
              Try again
            </Button>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}
