"use client"

import { useState, useTransition } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import {
  IconRobot,
  IconCheck,
  IconAlertCircle,
  IconSearch,
  IconMessageCircle,
  IconDownload,
  IconLoader2,
  IconUsers,
} from "@tabler/icons-react"

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

interface MarketplaceClientProps {
  templates: AgentTemplate[]
  acquiredTemplateIds: Set<string>
  connectedProviders: Set<string>
  categories: string[]
}

export function MarketplaceClient({
  templates,
  acquiredTemplateIds,
  connectedProviders,
  categories,
}: MarketplaceClientProps) {
  const [search, setSearch] = useState("")
  const [selectedCategory, setSelectedCategory] = useState("all")
  const [acquiredIds, setAcquiredIds] = useState<Set<string>>(new Set(acquiredTemplateIds))
  const [acquiring, setAcquiring] = useState<string | null>(null)
  const [, startTransition] = useTransition()

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
        startTransition(() => {
          setAcquiredIds((prev) => new Set([...prev, templateId]))
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
            const isAcquiring = acquiring === template.id
            const missing = template.required_integrations.filter(
              (p) => !connectedProviders.has(p)
            )

            return (
              <Card key={template.id} className="flex flex-col">
                <CardHeader>
                  <div className="flex items-start gap-3">
                    <div className="flex size-10 shrink-0 items-center justify-center bg-muted">
                      <IconRobot className="size-5" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 flex-wrap">
                        <CardTitle className="text-base">{template.display_name}</CardTitle>
                        <Badge variant="secondary" className="capitalize text-xs">
                          {template.category}
                        </Badge>
                      </div>
                      <CardDescription className="mt-0.5">{template.description}</CardDescription>
                    </div>
                  </div>
                </CardHeader>
                <CardContent className="flex flex-col gap-3 flex-1">
                  {/* Tags */}
                  {template.tags.length > 0 && (
                    <div className="flex flex-wrap gap-1">
                      {template.tags.map((tag) => (
                        <Badge key={tag} variant="outline" className="text-xs">
                          {tag}
                        </Badge>
                      ))}
                    </div>
                  )}

                  {/* Required integrations */}
                  {template.required_integrations.length > 0 && (
                    <div className="flex flex-col gap-1.5">
                      <p className="text-xs text-muted-foreground font-medium">Required integrations</p>
                      <div className="flex flex-wrap gap-1.5">
                        {template.required_integrations.map((provider) => {
                          const connected = connectedProviders.has(provider)
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
                  {!isAcquired && missing.length > 0 && (
                    <p className="text-xs text-muted-foreground">
                      Connect{" "}
                      <a href="/integrations" className="underline underline-offset-2">
                        {missing.join(", ")}
                      </a>{" "}
                      to use this agent.
                    </p>
                  )}

                  {/* Acquisition count */}
                  <div className="flex items-center gap-1.5 text-xs text-muted-foreground mt-auto">
                    <IconUsers className="size-3" />
                    <span>
                      {template.acquisition_count > 0
                        ? `${template.acquisition_count} user${template.acquisition_count === 1 ? "" : "s"}`
                        : "Be the first"}
                    </span>
                  </div>

                  {/* Action button */}
                  {isAcquired ? (
                    <Button asChild size="sm" className="w-full">
                      <a href={`/chat/${template.name}`}>
                        <IconMessageCircle className="size-4" />
                        Open
                      </a>
                    </Button>
                  ) : (
                    <Button
                      size="sm"
                      className="w-full"
                      onClick={() => handleAcquire(template.id)}
                      disabled={isAcquiring}
                    >
                      {isAcquiring ? (
                        <IconLoader2 className="size-4 animate-spin" />
                      ) : (
                        <IconDownload className="size-4" />
                      )}
                      {isAcquiring ? "Getting…" : "Get"}
                    </Button>
                  )}
                </CardContent>
              </Card>
            )
          })}
        </div>
      )}
    </div>
  )
}
