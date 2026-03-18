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
import { MarketplaceClient } from "./marketplace-client"

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
  status: "pending" | "approved" | "rejected"
}

interface UserIntegration {
  provider: string
  status: string
}

export default async function MarketplacePage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user) {
    redirect("/login")
  }

  // Fetch all active templates ordered by popularity
  const { data: templates } = await supabase
    .from("agent_templates")
    .select("id, name, display_name, description, category, tags, required_integrations, acquisition_count, status")
    .eq("status", "active")
    .order("acquisition_count", { ascending: false })

  // Fetch user's acquired agents with status
  const { data: userAgents } = await supabase
    .from("user_agents")
    .select("template_id, status")
    .eq("user_id", user.id)

  // Fetch user's active integrations
  const { data: integrations } = await supabase
    .from("user_integrations")
    .select("provider, status")
    .eq("user_id", user.id)
    .eq("status", "active")

  const acquiredTemplateIds = new Set(
    (userAgents ?? []).map((a: UserAgent) => a.template_id)
  )

  // Map template_id → acquisition status for multi-state UI
  const acquisitionStatuses = Object.fromEntries(
    (userAgents ?? []).map((a: UserAgent) => [a.template_id, a.status])
  )

  const userIsAdmin = isAdmin(user.email)

  const connectedProviders = new Set(
    (integrations ?? []).map((i: UserIntegration) => i.provider)
  )

  const templateList = (templates ?? []) as AgentTemplate[]

  // Derive unique categories
  const categories = Array.from(
    new Set(templateList.map((t) => t.category))
  ).filter(Boolean).sort()

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
                <BreadcrumbPage>Marketplace</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-6 p-6">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Marketplace</h1>
            <p className="text-sm text-muted-foreground">
              Discover and add AI agents to your workspace.
            </p>
          </div>

          <MarketplaceClient
            templates={templateList}
            acquiredTemplateIds={acquiredTemplateIds}
            acquisitionStatuses={acquisitionStatuses}
            connectedProviders={connectedProviders}
            categories={categories}
          />
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
