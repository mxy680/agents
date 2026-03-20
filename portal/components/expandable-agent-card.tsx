"use client"

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import {
  IconRobot,
  IconCheck,
  IconAlertCircle,
  IconMessageCircle,
  IconDownload,
  IconLoader2,
  IconClock,
} from "@tabler/icons-react"

interface Job {
  id: string
  displayName: string
  description: string
  schedule: string
}

interface ExpandableAgentCardProps {
  // Agent identity
  templateId: string
  name: string
  displayName: string
  description: string
  requiredIntegrations: string[]
  connectedProviders: string[]

  // Acquisition state (null = not acquired, used in marketplace)
  acquisitionStatus: "pending" | "approved" | "rejected" | null
  reviewerNote?: string | null

  // Jobs for this agent
  jobs: Job[]

  // Variant
  variant: "agents" | "marketplace"

  // Marketplace callbacks
  onAcquire?: (templateId: string) => void
  acquiring?: boolean
}

export function ExpandableAgentCard({
  templateId,
  name,
  displayName,
  description,
  requiredIntegrations,
  connectedProviders,
  acquisitionStatus,
  reviewerNote,
  jobs,
  variant,
  onAcquire,
  acquiring = false,
}: ExpandableAgentCardProps) {
  const connectedSet = new Set(connectedProviders)
  const missing = requiredIntegrations.filter((p) => !connectedSet.has(p))

  // Agents variant state
  const isApproved = acquisitionStatus === "approved"
  const isPending = acquisitionStatus === "pending"
  const isRejected = acquisitionStatus === "rejected"
  const canChat = isApproved && missing.length === 0

  // Marketplace variant state
  const isAcquired = acquisitionStatus !== null
  const isAcquiredApproved = isAcquired && acquisitionStatus === "approved"
  const isAcquiredPending = isAcquired && acquisitionStatus === "pending"
  const isAcquiredRejected = isAcquired && acquisitionStatus === "rejected"

  function renderAgentsButton() {
    if (isPending) {
      return (
        <Button size="sm" className="w-full" disabled>
          <IconMessageCircle className="size-4" />
          Awaiting Approval
        </Button>
      )
    }
    if (isRejected) {
      return (
        <Button size="sm" className="w-full" variant="outline" disabled>
          <IconMessageCircle className="size-4" />
          Access Denied
        </Button>
      )
    }
    return (
      <Button asChild={canChat} disabled={!canChat} size="sm" className="w-full">
        {canChat ? (
          <a href={`/chat/${name}`}>
            <IconMessageCircle className="size-4" />
            Chat
          </a>
        ) : (
          <>
            <IconMessageCircle className="size-4" />
            Chat
          </>
        )}
      </Button>
    )
  }

  function renderMarketplaceButton() {
    if (isAcquiredPending) {
      return (
        <Button
          size="sm"
          className="w-full bg-yellow-500/20 text-yellow-400 border border-yellow-500/30 hover:bg-yellow-500/20 cursor-default"
          disabled
        >
          Pending Approval
        </Button>
      )
    }
    if (isAcquiredRejected) {
      return (
        <Button size="sm" className="w-full" variant="outline" disabled>
          Access Denied
        </Button>
      )
    }
    if (isAcquiredApproved) {
      return (
        <Button asChild size="sm" className="w-full">
          <a href={`/chat/${name}`}>
            <IconMessageCircle className="size-4" />
            Open Chat
          </a>
        </Button>
      )
    }
    return (
      <Button
        size="sm"
        className="w-full"
        onClick={() => onAcquire?.(templateId)}
        disabled={acquiring}
      >
        {acquiring ? (
          <IconLoader2 className="size-4 animate-spin" />
        ) : (
          <IconDownload className="size-4" />
        )}
        {acquiring ? "Getting…" : "Get for Free"}
      </Button>
    )
  }

  return (
    <Card className="flex flex-col">
      <CardHeader>
        <div className="flex items-start gap-3">
          <div className="flex size-10 shrink-0 items-center justify-center bg-muted rounded-md">
            <IconRobot className="size-5" />
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 flex-wrap">
              <CardTitle className="text-base">{displayName}</CardTitle>
              {variant === "agents" && isPending && (
                <Badge variant="outline" className="text-xs bg-yellow-500/20 text-yellow-400 border-yellow-500/30">
                  Awaiting Approval
                </Badge>
              )}
              {variant === "agents" && isRejected && (
                <Badge variant="outline" className="text-xs bg-red-500/20 text-red-400 border-red-500/30">
                  Access Denied
                </Badge>
              )}
            </div>
          </div>
        </div>
      </CardHeader>

      <CardContent className="flex flex-1 flex-col gap-3">
        <CardDescription className="flex-1">{description}</CardDescription>

        {/* Required integrations with status */}
        {requiredIntegrations.length > 0 && (
          <div>
            <p className="text-xs font-medium text-muted-foreground mb-1.5">Required integrations</p>
            <div className="flex flex-wrap gap-1.5">
              {requiredIntegrations.map((provider) => {
                const connected = connectedSet.has(provider)
                return (
                  <Badge
                    key={provider}
                    variant={connected ? "outline" : "destructive"}
                    className="gap-1 capitalize"
                  >
                    {connected ? (
                      <IconCheck className="size-2.5" />
                    ) : (
                      <IconAlertCircle className="size-2.5" />
                    )}
                    {provider}
                  </Badge>
                )
              })}
            </div>
          </div>
        )}

        {/* Missing integrations hint */}
        {variant === "agents" && isApproved && !canChat && (
          <p className="text-xs text-muted-foreground mt-2">
            Connect{" "}
            <a href="/integrations" className="underline underline-offset-2">
              {missing.join(", ")}
            </a>{" "}
            to use this agent.
          </p>
        )}
        {variant === "marketplace" && !isAcquired && missing.length > 0 && (
          <p className="text-xs text-muted-foreground mt-2">
            Connect{" "}
            <a href="/integrations" className="underline underline-offset-2">
              {missing.join(", ")}
            </a>{" "}
            to use this agent.
          </p>
        )}

        {/* Reviewer note */}
        {isRejected && reviewerNote && (
          <p className="text-xs text-muted-foreground mt-2">
            Reason: {reviewerNote}
          </p>
        )}

        {/* Jobs section */}
        {jobs.length > 0 && (
          <div className="mt-3">
            <p className="text-xs font-medium text-muted-foreground mb-1.5">Scheduled Jobs</p>
            <div className="space-y-1.5">
              {jobs.map((job) => (
                <div
                  key={job.id}
                  className="flex items-center justify-between text-xs bg-muted/40 rounded px-2.5 py-1.5"
                >
                  <div className="flex items-center gap-1.5 min-w-0">
                    <IconClock className="size-3 shrink-0 text-muted-foreground" />
                    <span className="font-medium truncate">{job.displayName}</span>
                  </div>
                  <span className="text-muted-foreground ml-2 shrink-0">{job.schedule}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Action button */}
        {variant === "agents" ? renderAgentsButton() : renderMarketplaceButton()}
      </CardContent>
    </Card>
  )
}
