"use client"

import { useRouter } from "next/navigation"
import { useState } from "react"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { IconX } from "@tabler/icons-react"

interface AccountItemProps {
  id: string
  label: string
  status: string
}

export function AccountItem({ id, label, status }: AccountItemProps) {
  const router = useRouter()
  const [removing, setRemoving] = useState(false)

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

  return (
    <div className="flex items-center justify-between gap-2 border p-2 text-sm">
      <span className="truncate font-medium">{label}</span>
      <div className="flex items-center gap-2">
        <Badge variant={status === "active" ? "default" : "secondary"}>
          {status}
        </Badge>
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
  )
}
