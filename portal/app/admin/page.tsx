import { redirect } from "next/navigation"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
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
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import {
  IconPlugConnected,
  IconRobot,
  IconUsers,
  IconCalendarEvent,
} from "@tabler/icons-react"

export default async function AdminPage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user) {
    redirect("/login")
  }

  if (!isAdmin(user.email)) {
    redirect("/")
  }

  const admin = createAdminClient()

  // Fetch counts for dashboard
  const [integrationsRes, templatesRes, clientsRes, jobsRes] = await Promise.all([
    admin.from("user_integrations").select("id", { count: "exact", head: true }).eq("status", "active"),
    admin.from("agent_templates").select("id", { count: "exact", head: true }).eq("status", "active"),
    admin.from("clients").select("id", { count: "exact", head: true }).eq("active", true),
    admin.from("job_definitions").select("id", { count: "exact", head: true }).eq("enabled", true),
  ])

  const stats = [
    {
      title: "Active Integrations",
      count: integrationsRes.count ?? 0,
      icon: IconPlugConnected,
      href: "/integrations",
    },
    {
      title: "Agent Templates",
      count: templatesRes.count ?? 0,
      icon: IconRobot,
      href: "/chat",
    },
    {
      title: "Clients",
      count: clientsRes.count ?? 0,
      icon: IconUsers,
      href: "/admin/clients",
    },
    {
      title: "Active Jobs",
      count: jobsRes.count ?? 0,
      icon: IconCalendarEvent,
      href: "/jobs",
    },
  ]

  return (
    <SidebarProvider>
      <AppSidebar />
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
                <BreadcrumbPage>Admin</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-6 p-6">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Dashboard</h1>
            <p className="text-sm text-muted-foreground">
              Overview of integrations, agents, and clients.
            </p>
          </div>

          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
            {stats.map((stat) => (
              <a key={stat.title} href={stat.href}>
                <Card className="transition-colors hover:bg-muted/50">
                  <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle className="text-sm font-medium">{stat.title}</CardTitle>
                    <stat.icon className="size-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">{stat.count}</div>
                  </CardContent>
                </Card>
              </a>
            ))}
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
