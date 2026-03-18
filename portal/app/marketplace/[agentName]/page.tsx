import { createClient } from "@/lib/supabase/server"
import { redirect, notFound } from "next/navigation"
import { AppSidebar } from "@/components/app-sidebar"
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb"
import { Separator } from "@/components/ui/separator"
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar"
import { Badge } from "@/components/ui/badge"
import { IconRobot, IconCheck, IconAlertCircle, IconArrowLeft, IconUsers } from "@tabler/icons-react"
import { AgentDetailActions } from "./agent-detail-actions"

interface AgentTemplate {
  id: string
  name: string
  display_name: string
  description: string
  category: string
  tags: string[]
  required_integrations: string[]
  acquisition_count: number
  status: string
}

interface UserAgent {
  template_id: string
}

interface UserIntegration {
  provider: string
  status: string
}

export default async function AgentDetailPage({
  params,
}: {
  params: Promise<{ agentName: string }>
}) {
  const { agentName } = await params
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user) {
    redirect("/login")
  }

  // Fetch template by name
  const { data: template } = await supabase
    .from("agent_templates")
    .select("id, name, display_name, description, category, tags, required_integrations, acquisition_count, status")
    .eq("name", agentName)
    .single()

  if (!template || template.status !== "active") {
    notFound()
  }

  const t = template as AgentTemplate

  // Fetch user acquisition status
  const { data: userAgent } = await supabase
    .from("user_agents")
    .select("template_id")
    .eq("user_id", user.id)
    .eq("template_id", t.id)
    .maybeSingle()

  const isAcquired = !!(userAgent as UserAgent | null)

  // Fetch user integrations
  const { data: integrations } = await supabase
    .from("user_integrations")
    .select("provider, status")
    .eq("user_id", user.id)
    .eq("status", "active")

  const connectedProviders = new Set(
    (integrations ?? []).map((i: UserIntegration) => i.provider)
  )

  const missing = t.required_integrations.filter((p) => !connectedProviders.has(p))

  return (
    <SidebarProvider>
      <AppSidebar
        user={{
          email: user.email ?? undefined,
          name: user.user_metadata?.full_name ?? user.user_metadata?.name,
        }}
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
                <BreadcrumbLink href="/marketplace">Marketplace</BreadcrumbLink>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbPage>{t.display_name}</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-8 p-6 max-w-2xl">
          {/* Back link */}
          <a
            href="/marketplace"
            className="flex items-center gap-1.5 text-xs text-muted-foreground hover:text-foreground transition-colors w-fit"
          >
            <IconArrowLeft className="size-3.5" />
            Back to Marketplace
          </a>

          {/* Agent header */}
          <div className="flex items-start gap-4">
            <div className="flex size-16 shrink-0 items-center justify-center bg-muted">
              <IconRobot className="size-8" />
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 flex-wrap">
                <h1 className="text-2xl font-semibold tracking-tight">{t.display_name}</h1>
                <Badge variant="secondary" className="capitalize">{t.category}</Badge>
              </div>
              <p className="mt-1 text-sm text-muted-foreground">{t.description}</p>
              <div className="flex items-center gap-1.5 mt-2 text-xs text-muted-foreground">
                <IconUsers className="size-3.5" />
                <span>
                  {t.acquisition_count > 0
                    ? `${t.acquisition_count} user${t.acquisition_count === 1 ? "" : "s"}`
                    : "Be the first to add this agent"}
                </span>
              </div>
            </div>
          </div>

          {/* Tags */}
          {t.tags.length > 0 && (
            <div className="flex flex-col gap-2">
              <h2 className="text-sm font-medium">Tags</h2>
              <div className="flex flex-wrap gap-1.5">
                {t.tags.map((tag) => (
                  <Badge key={tag} variant="outline">{tag}</Badge>
                ))}
              </div>
            </div>
          )}

          {/* Required integrations */}
          {t.required_integrations.length > 0 && (
            <div className="flex flex-col gap-2">
              <h2 className="text-sm font-medium">Required integrations</h2>
              <div className="flex flex-wrap gap-2">
                {t.required_integrations.map((provider) => {
                  const connected = connectedProviders.has(provider)
                  return (
                    <Badge
                      key={provider}
                      variant={connected ? "outline" : "destructive"}
                      className="gap-1.5 capitalize text-sm px-3 py-1"
                    >
                      {connected ? (
                        <IconCheck className="size-3" />
                      ) : (
                        <IconAlertCircle className="size-3" />
                      )}
                      {provider}
                      <span className="text-xs opacity-70">
                        {connected ? "Connected" : "Not connected"}
                      </span>
                    </Badge>
                  )
                })}
              </div>
              {missing.length > 0 && (
                <p className="text-xs text-muted-foreground">
                  You need to{" "}
                  <a href="/integrations" className="underline underline-offset-2">
                    connect {missing.join(", ")}
                  </a>{" "}
                  before chatting with this agent.
                </p>
              )}
            </div>
          )}

          {/* Action button */}
          <AgentDetailActions
            templateId={t.id}
            agentName={t.name}
            isAcquired={isAcquired}
          />
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
