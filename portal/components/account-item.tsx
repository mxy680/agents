"use client"

import { useRouter } from "next/navigation"
import { useState } from "react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { IconX, IconPlayerPlay, IconCheck, IconAlertTriangle, IconLoader2 } from "@tabler/icons-react"

interface AccountItemProps {
  id: string
  label: string
  status: string
}

export function AccountItem({ id, label, status }: AccountItemProps) {
  const router = useRouter()
  const [removing, setRemoving] = useState(false)
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState<"pass" | "fail" | null>(null)
  const [testError, setTestError] = useState("")

  async function handleDisconnect() {
    setRemoving(true)
    const res = await fetch("/api/integrations/disconnect", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ id }),
    })
    if (res.ok) {
      router.refresh()
    } else {
      setRemoving(false)
    }
  }

  async function handleTest() {
    setTesting(true)
    setTestResult(null)
    setTestError("")
    try {
      const res = await fetch("/api/integrations/test", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ id }),
      })
      const data = await res.json()
      if (data.ok) {
        setTestResult("pass")
      } else {
        setTestResult("fail")
        setTestError(data.error || "Test failed")
      }
    } catch {
      setTestResult("fail")
      setTestError("Network error")
    } finally {
      setTesting(false)
      setTimeout(() => {
        setTestResult(null)
        setTestError("")
      }, 8000)
    }
  }

  return (
    <div className="flex flex-col gap-1">
      <div className="flex items-center justify-between gap-2 border p-2 text-sm">
        <span className="truncate font-medium">{label}</span>
        <div className="flex items-center gap-1">
          <Badge variant={status === "active" ? "default" : "secondary"}>
            {status}
          </Badge>
          <Button
            variant="ghost"
            size="icon"
            className="size-6"
            aria-label="Test credentials"
            onClick={handleTest}
            disabled={testing || removing}
          >
            {testing ? (
              <IconLoader2 className="size-3 animate-spin" />
            ) : testResult === "pass" ? (
              <IconCheck className="size-3 text-green-500" />
            ) : testResult === "fail" ? (
              <IconAlertTriangle className="size-3 text-red-500" />
            ) : (
              <IconPlayerPlay className="size-3" />
            )}
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="size-6"
            aria-label="Disconnect account"
            onClick={handleDisconnect}
            disabled={removing}
          >
            <IconX className="size-3" />
          </Button>
        </div>
      </div>
      {testError && (
        <p className="px-2 text-xs text-red-500 font-mono truncate" title={testError}>
          {testError}
        </p>
      )}
    </div>
  )
}
