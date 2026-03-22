import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { redirect } from "next/navigation"
import { isAdmin } from "@/lib/admin"
import { AppSidebar } from "@/components/app-sidebar"
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
import { IconRobot, IconMessage } from "@tabler/icons-react"

export default async function AgentsPage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user) {
    redirect("/login")
  }

  if (!isAdmin(user.email)) {
    redirect("/")
  }

  const admin = createAdminClient()

  const [templatesRes, integrationsRes, clientAgentsRes, convsRes] = await Promise.all([
    admin
      .from("agent_templates")
      .select("id, name, display_name, description, required_integrations, status")
      .order("display_name", { ascending: true }),
    admin
      .from("user_integrations")
      .select("provider")
      .eq("status", "active"),
    admin
      .from("client_agents")
      .select("template_id"),
    admin
      .from("conversations")
      .select("agent_name"),
  ])

  const templates = templatesRes.data ?? []
  const connectedProviders = new Set((integrationsRes.data ?? []).map((i) => i.provider))

  const clientCountMap: Record<string, number> = {}
  for (const row of clientAgentsRes.data ?? []) {
    clientCountMap[row.template_id] = (clientCountMap[row.template_id] ?? 0) + 1
  }

  const convCountMap: Record<string, number> = {}
  for (const row of convsRes.data ?? []) {
    convCountMap[row.agent_name] = (convCountMap[row.agent_name] ?? 0) + 1
  }

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
                <BreadcrumbPage>Agents</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-6 p-6">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Agents</h1>
            <p className="text-sm text-muted-foreground">
              Manage agent templates, clients, and conversations.
            </p>
          </div>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {templates.map((template) => {
              const requiredIntegrations = (template.required_integrations ?? []) as string[]
              const clientCount = clientCountMap[template.id] ?? 0
              const conversationCount = convCountMap[template.name] ?? 0

              return (
                <Card key={template.id} className="h-full">
                  <CardHeader>
                    <div className="flex items-center gap-3">
                      <div className="flex size-10 items-center justify-center bg-muted">
                        <IconRobot className="size-5" />
                      </div>
                      <div className="flex-1">
                        <div className="flex items-center gap-2">
                          <CardTitle className="text-base">
                            <a href={`/agents/${template.name}`} className="hover:underline">
                              {template.display_name}
                            </a>
                          </CardTitle>
                          <Badge variant={template.status === "active" ? "default" : "secondary"}>
                            {template.status}
                          </Badge>
                        </div>
                        <CardDescription>{template.description}</CardDescription>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className="flex flex-col gap-3">
                    {requiredIntegrations.length > 0 && (
                      <div className="flex flex-wrap gap-1.5">
                        {requiredIntegrations.map((provider) => (
                          <span
                            key={provider}
                            className="inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-xs"
                          >
                            <span
                              className={
                                connectedProviders.has(provider)
                                  ? "text-green-500"
                                  : "text-red-500"
                              }
                            >
                              ●
                            </span>
                            {provider}
                          </span>
                        ))}
                      </div>
                    )}
                    <p className="text-xs text-muted-foreground">
                      {clientCount} {clientCount === 1 ? "client" : "clients"} · {conversationCount}{" "}
                      {conversationCount === 1 ? "conversation" : "conversations"}
                    </p>
                    <div className="flex gap-2">
                      <Button size="sm" asChild>
                        <a href={`/chat/${template.name}`}>
                          <IconMessage className="size-4" />
                          Chat
                        </a>
                      </Button>
                      <Button size="sm" variant="outline" asChild>
                        <a href={`/agents/${template.name}`}>
                          Details
                        </a>
                      </Button>
                    </div>
                  </CardContent>
                </Card>
              )
            })}
            {templates.length === 0 && (
              <div className="col-span-3 flex flex-col items-center justify-center gap-2 py-16 text-muted-foreground">
                <IconRobot className="size-10" />
                <p>No agent templates found.</p>
              </div>
            )}
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
