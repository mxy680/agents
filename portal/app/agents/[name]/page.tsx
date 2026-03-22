import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { redirect, notFound } from "next/navigation"
import { isAdmin } from "@/lib/admin"
import { AppSidebar } from "@/components/app-sidebar"
import {
  Breadcrumb,
  BreadcrumbItem,
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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import { IconMessage } from "@tabler/icons-react"
import { AgentClients } from "./agent-clients"

export default async function AgentDetailPage({
  params,
}: {
  params: Promise<{ name: string }>
}) {
  const { name } = await params

  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user) {
    redirect("/login")
  }

  if (!isAdmin(user.email)) {
    redirect("/")
  }

  const admin = createAdminClient()

  // Fetch the template
  const { data: template } = await admin
    .from("agent_templates")
    .select("id, name, display_name, description, required_integrations, status")
    .eq("name", name)
    .single()

  if (!template) {
    notFound()
  }

  const requiredIntegrations = (template.required_integrations ?? []) as string[]

  const [
    integrationsRes,
    assignmentsRes,
    allClientsRes,
    conversationsRes,
    jobDefsRes,
  ] = await Promise.all([
    // Integration health
    admin
      .from("user_integrations")
      .select("provider")
      .eq("status", "active"),
    // Assigned clients
    admin
      .from("client_agents")
      .select("id, client_id, notes, created_at, clients(id, name, email, active)")
      .eq("template_id", template.id)
      .order("created_at", { ascending: false }),
    // All active clients (for assign dropdown)
    admin
      .from("clients")
      .select("id, name, email")
      .eq("active", true)
      .order("name", { ascending: true }),
    // Recent conversations
    admin
      .from("conversations")
      .select("id, title, client_id, created_at, clients(name)")
      .eq("agent_name", template.name)
      .order("created_at", { ascending: false })
      .limit(10),
    // Job definitions with last run
    admin
      .from("job_definitions")
      .select("id, slug, display_name, schedule, enabled")
      .eq("template_id", template.id)
      .order("display_name", { ascending: true }),
  ])

  const connectedProviders = new Set((integrationsRes.data ?? []).map((i) => i.provider))
  const assignments = (assignmentsRes.data ?? []) as unknown as Array<{
    id: string
    client_id: string
    notes: string | null
    created_at: string
    clients: { id: string; name: string; email: string | null; active: boolean } | null
  }>
  const allClients = (allClientsRes.data ?? []) as Array<{
    id: string
    name: string
    email: string | null
  }>
  const conversations = conversationsRes.data ?? []
  const jobDefs = jobDefsRes.data ?? []

  // Fetch last run status for each job definition
  let lastRunMap: Record<string, { status: string; completed_at: string | null }> = {}
  if (jobDefs.length > 0) {
    const jobDefIds = jobDefs.map((j) => j.id)
    const { data: lastRuns } = await admin
      .from("job_runs")
      .select("job_definition_id, status, completed_at")
      .in("job_definition_id", jobDefIds)
      .order("created_at", { ascending: false })

    for (const run of lastRuns ?? []) {
      if (!lastRunMap[run.job_definition_id]) {
        lastRunMap[run.job_definition_id] = {
          status: run.status,
          completed_at: run.completed_at,
        }
      }
    }
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
                <a href="/agents">Agents</a>
              </BreadcrumbItem>
              <BreadcrumbSeparator />
              <BreadcrumbItem>
                <BreadcrumbPage>{template.display_name}</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-6 p-6">
          {/* Header */}
          <div className="flex items-start justify-between gap-4">
            <div className="flex flex-col gap-1">
              <div className="flex items-center gap-2">
                <h1 className="text-2xl font-semibold tracking-tight">{template.display_name}</h1>
                <Badge variant={template.status === "active" ? "default" : "secondary"}>
                  {template.status}
                </Badge>
              </div>
              {template.description && (
                <p className="text-sm text-muted-foreground">{template.description}</p>
              )}
            </div>
            <Button asChild>
              <a href={`/chat/${template.name}`}>
                <IconMessage className="size-4" />
                New Chat
              </a>
            </Button>
          </div>

          {/* Integrations */}
          {requiredIntegrations.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Required Integrations</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="flex flex-col gap-2">
                  {requiredIntegrations.map((provider) => {
                    const connected = connectedProviders.has(provider)
                    return (
                      <div key={provider} className="flex items-center gap-2 text-sm">
                        <span className={connected ? "text-green-500" : "text-red-500"}>●</span>
                        <span className="capitalize">{provider}</span>
                        <span className="text-muted-foreground">
                          {connected ? "Connected" : "Not connected"}
                        </span>
                      </div>
                    )
                  })}
                </div>
              </CardContent>
            </Card>
          )}

          {/* Clients */}
          <Card>
            <CardContent className="pt-6">
              <AgentClients
                agentName={template.name}
                initialAssignments={assignments}
                availableClients={allClients}
              />
            </CardContent>
          </Card>

          {/* Recent Conversations */}
          <Card>
            <CardHeader>
              <CardTitle className="text-base">Recent Conversations</CardTitle>
            </CardHeader>
            <CardContent>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Title</TableHead>
                    <TableHead>Client</TableHead>
                    <TableHead>Date</TableHead>
                    <TableHead />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {conversations.length === 0 && (
                    <TableRow>
                      <TableCell colSpan={4} className="text-center text-muted-foreground">
                        No conversations yet.
                      </TableCell>
                    </TableRow>
                  )}
                  {conversations.map((conv) => {
                    const clientData = conv.clients as unknown as { name: string } | null
                    return (
                      <TableRow key={conv.id}>
                        <TableCell className="font-medium">
                          {conv.title ?? "Untitled"}
                        </TableCell>
                        <TableCell className="text-muted-foreground">
                          {clientData?.name ?? "—"}
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {new Date(conv.created_at).toLocaleDateString()}
                        </TableCell>
                        <TableCell>
                          <a
                            href={`/chat/${template.name}?conversation=${conv.id}`}
                            className="text-sm text-primary hover:underline"
                          >
                            Open
                          </a>
                        </TableCell>
                      </TableRow>
                    )
                  })}
                </TableBody>
              </Table>
            </CardContent>
          </Card>

          {/* Jobs */}
          {jobDefs.length > 0 && (
            <Card>
              <CardHeader>
                <CardTitle className="text-base">Scheduled Jobs</CardTitle>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Job</TableHead>
                      <TableHead>Schedule</TableHead>
                      <TableHead>Status</TableHead>
                      <TableHead>Last Run</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {jobDefs.map((job) => {
                      const lastRun = lastRunMap[job.id]
                      return (
                        <TableRow key={job.id}>
                          <TableCell className="font-medium">{job.display_name}</TableCell>
                          <TableCell className="font-mono text-xs text-muted-foreground">
                            {job.schedule}
                          </TableCell>
                          <TableCell>
                            <Badge variant={job.enabled ? "default" : "secondary"}>
                              {job.enabled ? "Enabled" : "Disabled"}
                            </Badge>
                          </TableCell>
                          <TableCell className="text-sm text-muted-foreground">
                            {lastRun ? (
                              <span className="flex items-center gap-1.5">
                                <Badge
                                  variant={
                                    lastRun.status === "completed"
                                      ? "default"
                                      : lastRun.status === "failed" || lastRun.status === "timed_out"
                                      ? "destructive"
                                      : "secondary"
                                  }
                                >
                                  {lastRun.status}
                                </Badge>
                                {lastRun.completed_at && (
                                  <span>{new Date(lastRun.completed_at).toLocaleDateString()}</span>
                                )}
                              </span>
                            ) : (
                              "Never"
                            )}
                          </TableCell>
                        </TableRow>
                      )
                    })}
                  </TableBody>
                </Table>
              </CardContent>
            </Card>
          )}
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
