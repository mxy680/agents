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
import { Button } from "@/components/ui/button"
import { IconBuildingStore } from "@tabler/icons-react"
import { toHumanReadable } from "@/lib/cron"
import { AgentsClient } from "./agents-client"

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

interface JobDefinition {
  id: string
  template_id: string
  display_name: string
  description: string
  schedule: string
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

  const connectedProviders = (integrations ?? []).map((i: UserIntegration) => i.provider)

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

  // Fetch job definitions for all acquired template IDs
  const templateIds = acquiredTemplates.map((t) => t.id)
  const jobsByTemplate: Record<string, Array<{ id: string; displayName: string; description: string; schedule: string }>> = {}

  if (templateIds.length > 0) {
    const { data: jobDefs } = await supabase
      .from("job_definitions")
      .select("id, template_id, display_name, description, schedule")
      .in("template_id", templateIds)
      .eq("enabled", true)

    for (const job of (jobDefs ?? []) as JobDefinition[]) {
      if (!jobsByTemplate[job.template_id]) {
        jobsByTemplate[job.template_id] = []
      }
      jobsByTemplate[job.template_id].push({
        id: job.id,
        displayName: job.display_name,
        description: job.description,
        schedule: toHumanReadable(job.schedule),
      })
    }
  }

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

          <AgentsClient
            templates={acquiredTemplates}
            connectedProviders={connectedProviders}
            jobsByTemplate={jobsByTemplate}
          />
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
