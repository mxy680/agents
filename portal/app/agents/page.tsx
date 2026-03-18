import { createClient } from "@/lib/supabase/server"
import { redirect } from "next/navigation"
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
import { IconRobot, IconMessageCircle, IconCheck, IconAlertCircle } from "@tabler/icons-react"

interface AgentTemplate {
  id: string
  name: string
  display_name: string
  description: string
  required_integrations: string[]
  status: string
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

  // Fetch active agent templates
  const { data: templates } = await supabase
    .from("agent_templates")
    .select("id, name, display_name, description, required_integrations, status")
    .eq("status", "active")
    .order("display_name")

  // Fetch user's active integrations
  const { data: integrations } = await supabase
    .from("user_integrations")
    .select("provider, status")
    .eq("user_id", user.id)
    .eq("status", "active")

  const connectedProviders = new Set(
    (integrations ?? []).map((i: UserIntegration) => i.provider)
  )

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
              Chat with AI agents that use your connected integrations.
            </p>
          </div>

          {(!templates || templates.length === 0) ? (
            <div className="flex flex-col items-center justify-center gap-3 py-16 text-center">
              <IconRobot className="size-12 text-muted-foreground/50" />
              <p className="text-sm text-muted-foreground">No agents available yet.</p>
            </div>
          ) : (
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {(templates as AgentTemplate[]).map((template) => {
                const missing = template.required_integrations.filter(
                  (p) => !connectedProviders.has(p)
                )
                const canChat = missing.length === 0

                return (
                  <Card key={template.id}>
                    <CardHeader>
                      <div className="flex items-start gap-3">
                        <div className="flex size-10 shrink-0 items-center justify-center bg-muted">
                          <IconRobot className="size-5" />
                        </div>
                        <div className="flex-1 min-w-0">
                          <CardTitle className="text-base">{template.display_name}</CardTitle>
                          <CardDescription className="mt-0.5">{template.description}</CardDescription>
                        </div>
                      </div>
                    </CardHeader>
                    <CardContent className="flex flex-col gap-3">
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

                      {!canChat && (
                        <p className="text-xs text-muted-foreground">
                          Connect{" "}
                          <a href="/integrations" className="underline underline-offset-2">
                            {missing.join(", ")}
                          </a>{" "}
                          to use this agent.
                        </p>
                      )}

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
