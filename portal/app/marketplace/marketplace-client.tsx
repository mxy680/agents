"use client"

import { useState, useTransition } from "react"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { ExpandableAgentCard } from "@/components/expandable-agent-card"
import { IconRobot, IconSearch } from "@tabler/icons-react"

interface AgentTemplate {
  id: string
  name: string
  display_name: string
  description: string
  category: string
  tags: string[]
  required_integrations: string[]
  acquisition_count: number
}

type AcquisitionStatus = "pending" | "approved" | "rejected"

interface JobRow {
  id: string
  displayName: string
  description: string
  schedule: string
}

interface MarketplaceClientProps {
  templates: AgentTemplate[]
  acquiredTemplateIds: Set<string>
  acquisitionStatuses: Record<string, AcquisitionStatus>
  connectedProviders: Set<string>
  categories: string[]
  jobsByTemplate: Record<string, JobRow[]>
}

export function MarketplaceClient({
  templates,
  acquiredTemplateIds,
  acquisitionStatuses,
  connectedProviders,
  categories,
  jobsByTemplate,
}: MarketplaceClientProps) {
  const [search, setSearch] = useState("")
  const [selectedCategory, setSelectedCategory] = useState("all")
  const [acquiredIds, setAcquiredIds] = useState<Set<string>>(new Set(acquiredTemplateIds))
  const [statuses, setStatuses] = useState<Record<string, AcquisitionStatus>>(acquisitionStatuses)
  const [acquiring, setAcquiring] = useState<string | null>(null)
  const [, startTransition] = useTransition()

  const connectedProvidersArray = Array.from(connectedProviders)

  const filtered = templates.filter((t) => {
    const matchesSearch =
      search.trim() === "" ||
      t.display_name.toLowerCase().includes(search.toLowerCase()) ||
      t.description.toLowerCase().includes(search.toLowerCase()) ||
      t.tags.some((tag) => tag.toLowerCase().includes(search.toLowerCase()))
    const matchesCategory =
      selectedCategory === "all" || t.category === selectedCategory
    return matchesSearch && matchesCategory
  })

  async function handleAcquire(templateId: string) {
    setAcquiring(templateId)
    try {
      const res = await fetch("/api/agents/acquire", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ templateId }),
      })
      if (res.ok) {
        const data = await res.json().catch(() => ({}))
        const newStatus: AcquisitionStatus = data.status ?? "pending"
        startTransition(() => {
          setAcquiredIds((prev) => new Set([...prev, templateId]))
          setStatuses((prev) => ({ ...prev, [templateId]: newStatus }))
        })
      }
    } finally {
      setAcquiring(null)
    }
  }

  return (
    <div className="flex flex-col gap-6">
      {/* Search and category filter */}
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
        <div className="relative flex-1">
          <IconSearch className="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
          <Input
            placeholder="Search agents…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="pl-8"
          />
        </div>
        <div className="flex flex-wrap gap-1.5">
          <Button
            variant={selectedCategory === "all" ? "default" : "outline"}
            size="sm"
            onClick={() => setSelectedCategory("all")}
          >
            All
          </Button>
          {categories.map((cat) => (
            <Button
              key={cat}
              variant={selectedCategory === cat ? "default" : "outline"}
              size="sm"
              onClick={() => setSelectedCategory(cat)}
              className="capitalize"
            >
              {cat}
            </Button>
          ))}
        </div>
      </div>

      {/* Grid */}
      {filtered.length === 0 ? (
        <div className="flex flex-col items-center justify-center gap-3 py-16 text-center">
          <IconRobot className="size-12 text-muted-foreground/50" />
          <p className="text-sm text-muted-foreground">No agents found.</p>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {filtered.map((template) => {
            const isAcquired = acquiredIds.has(template.id)
            const acquisitionStatus = isAcquired ? (statuses[template.id] ?? null) : null
            const isAcquiring = acquiring === template.id

            return (
              <ExpandableAgentCard
                key={template.id}
                templateId={template.id}
                name={template.name}
                displayName={template.display_name}
                description={template.description}
                requiredIntegrations={template.required_integrations}
                connectedProviders={connectedProvidersArray}
                acquisitionStatus={acquisitionStatus}
                jobs={jobsByTemplate[template.id] ?? []}
                variant="marketplace"
                onAcquire={handleAcquire}
                acquiring={isAcquiring}
              />
            )
          })}
        </div>
      )}
    </div>
  )
}
