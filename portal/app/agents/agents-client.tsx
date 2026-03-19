"use client"

import { Button } from "@/components/ui/button"
import { ExpandableAgentCard } from "@/components/expandable-agent-card"
import { IconRobot, IconBuildingStore } from "@tabler/icons-react"

interface JobRow {
  id: string
  displayName: string
  description: string
  schedule: string
}

interface TemplateRow {
  id: string
  name: string
  display_name: string
  description: string
  required_integrations: string[]
  acquisitionStatus: "pending" | "approved" | "rejected"
  reviewerNote: string | null
}

interface AgentsClientProps {
  templates: TemplateRow[]
  connectedProviders: string[]
  jobsByTemplate: Record<string, JobRow[]>
}

export function AgentsClient({ templates, connectedProviders, jobsByTemplate }: AgentsClientProps) {
  if (templates.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center gap-4 py-16 text-center">
        <IconRobot className="size-12 text-muted-foreground/50" />
        <div>
          <p className="text-sm font-medium">No agents yet.</p>
          <p className="text-sm text-muted-foreground mt-0.5">
            Browse the marketplace to get started.
          </p>
        </div>
        <Button asChild variant="outline">
          <a href="/marketplace">
            <IconBuildingStore className="size-4" />
            Browse Marketplace
          </a>
        </Button>
      </div>
    )
  }

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {templates.map((template) => (
        <ExpandableAgentCard
          key={template.id}
          templateId={template.id}
          name={template.name}
          displayName={template.display_name}
          description={template.description}
          requiredIntegrations={template.required_integrations}
          connectedProviders={connectedProviders}
          acquisitionStatus={template.acquisitionStatus}
          reviewerNote={template.reviewerNote}
          jobs={jobsByTemplate[template.id] ?? []}
          variant="agents"
        />
      ))}
    </div>
  )
}
