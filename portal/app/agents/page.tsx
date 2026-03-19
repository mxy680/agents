import { createClient } from "@/lib/supabase/server"
import { redirect } from "next/navigation"
import { AppSidebar } from "@/components/app-sidebar"
import { isAdmin } from "@/lib/admin"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { IconRobot, IconMessageCircle, IconCheck, IconAlertCircle, IconBuildingStore } from "@tabler/icons-react"

interface AgentTemplate {
  id: string
  name: string
  display_name: string
  description: string
  required_integrations: string[]
  status: string
}

interface UserAgent {
  template_id: string
  status: "pending" | "approved" | "rejected"
  reviewer_note: string | null
  agent_templates: AgentTemplate | AgentTemplate[]
}

interface AcquiredTemplate extends AgentTemplate {
  acquisitionStatus: "pending" | "approved" | "rejected"
  reviewerNote: string | null
}

interface UserIntegration {
  provider: string
  status: string
}

export default async function AgentsPage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user) {
    redirect("/login")
  }

  // Fetch user's acquired agents joined with template details
  const { data: userAgents } = await supabase
    .from("user_agents")
    .select("template_id, status, reviewer_note, agent_templates(id, name, display_name, description, required_integrations, status)")
    .eq("user_id", user.id)

  // Fetch user's active integrations
  const { data: integrations } = await supabase
    .from("user_integrations")
    .select("provider, status")
    .eq("user_id", user.id)
    .eq("status", "active")

  const connectedProviders = new Set(
    (integrations ?? []).map((i: UserIntegration) => i.provider)
  )

  // Map to active templates, carrying acquisition status
  // Supabase returns one-to-one FK joins as an object, cast accordingly
  const acquiredTemplates = (userAgents ?? [])
    .map((ua: UserAgent) => {
      const t = Array.isArray(ua.agent_templates) ? ua.agent_templates[0] : ua.agent_templates
      if (!t || t.status !== "active") return null
      return {
        ...t,
        acquisitionStatus: ua.status,
        reviewerNote: ua.reviewer_note ?? null,
      } as AcquiredTemplate
    })
    .filter((t): t is AcquiredTemplate => !!t)

  const userIsAdmin = isAdmin(user.email)

  return (
    <SidebarProvider>
      <AppSidebar
        user={{
          email: user.email ?? undefined,
          name: user.user_metadata?.full_name ?? user.user_metadata?.name,
        }}
        isAdmin={userIsAdmin}
      />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger className="-ml-1" />
          <Separator
            orientation="vertical"
            className="mr-2 data-vertical:h-4 data-vertical:self-auto"
          />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbPage>My Agents</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
          <div className="ml-auto">
            <Button asChild variant="outline" size="sm">
              <a href="/marketplace">
                <IconBuildingStore className="size-4" />
                Browse Marketplace
              </a>
            </Button>
          </div>
        </header>
        <div className="flex flex-1 flex-col gap-6 p-6">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">My Agents</h1>
            <p className="text-sm text-muted-foreground">
              Chat with AI agents that use your connected integrations.
            </p>
          </div>

          {acquiredTemplates.length === 0 ? (
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
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {acquiredTemplates.map((template) => {
                const missing = template.required_integrations.filter(
                  (p) => !connectedProviders.has(p)
                )
                const isApproved = template.acquisitionStatus === "approved"
                const isPending = template.acquisitionStatus === "pending"
                const isRejected = template.acquisitionStatus === "rejected"
                const canChat = isApproved && missing.length === 0

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
                            {isPending && (
                              <Badge variant="outline" className="text-xs bg-yellow-500/20 text-yellow-400 border-yellow-500/30">
                                Awaiting Approval
                              </Badge>
                            )}
                            {isRejected && (
                              <Badge variant="outline" className="text-xs bg-red-500/20 text-red-400 border-red-500/30">
                                Access Denied
                              </Badge>
                            )}
                          </div>
                        </div>
                      </div>
                    </CardHeader>
                    <CardContent className="flex flex-1 flex-col gap-3">
                      {/* Description grows to fill space, pushing integrations+button to bottom */}
                      <CardDescription className="flex-1">{template.description}</CardDescription>

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

                      {isApproved && !canChat && (
                        <p className="text-xs text-muted-foreground">
                          Connect{" "}
                          <a href="/integrations" className="underline underline-offset-2">
                            {missing.join(", ")}
                          </a>{" "}
                          to use this agent.
                        </p>
                      )}

                      {isRejected && template.reviewerNote && (
                        <p className="text-xs text-muted-foreground">
                          Reason: {template.reviewerNote}
                        </p>
                      )}

                      {isPending ? (
                        <Button size="sm" className="w-full" disabled>
                          <IconMessageCircle />
                          Awaiting Approval
                        </Button>
                      ) : isRejected ? (
                        <Button size="sm" className="w-full" disabled variant="outline">
                          <IconMessageCircle />
                          Access Denied
                        </Button>
                      ) : (
                        <Button
                          asChild={canChat}
                          disabled={!canChat}
                          size="sm"
                          className="w-full"
                        >
                          {canChat ? (
                            <a href={`/chat/${template.name}`}>
                              <IconMessageCircle />
                              Chat
                            </a>
                          ) : (
                            <>
                              <IconMessageCircle />
                              Chat
                            </>
                          )}
                        </Button>
                      )}
                    </CardContent>
                  </Card>
                )
              })}
            </div>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
