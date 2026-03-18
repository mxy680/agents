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
import { AdminTable } from "./admin-table"

interface ApprovalRequest {
  id: string
  user_id: string
  user_email: string
  template_id: string
  template_name: string
  template_display_name: string
  status: "pending" | "approved" | "rejected"
  acquired_at: string
  reviewed_at: string | null
  reviewer_note: string | null
}

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

  // Fetch all user_agents with template info
  const { data: rows } = await admin
    .from("user_agents")
    .select("id, user_id, template_id, status, acquired_at, reviewed_at, reviewer_note, agent_templates(id, name, display_name)")
    .order("acquired_at", { ascending: true })

  // Fetch emails for unique user IDs
  const userIds = Array.from(new Set((rows ?? []).map((r) => r.user_id)))
  const emailMap: Record<string, string> = {}

  for (const uid of userIds) {
    try {
      const { data: userData } = await admin.auth.admin.getUserById(uid)
      if (userData.user?.email) {
        emailMap[uid] = userData.user.email
      }
    } catch {
      // non-fatal
    }
  }

  const requests: ApprovalRequest[] = (rows ?? []).map((row) => {
    const tmpl = Array.isArray(row.agent_templates)
      ? row.agent_templates[0]
      : row.agent_templates
    return {
      id: row.id,
      user_id: row.user_id,
      user_email: emailMap[row.user_id] ?? "Unknown",
      template_id: row.template_id,
      template_name: (tmpl as { name?: string } | null)?.name ?? "",
      template_display_name: (tmpl as { display_name?: string } | null)?.display_name ?? "",
      status: row.status as "pending" | "approved" | "rejected",
      acquired_at: row.acquired_at,
      reviewed_at: row.reviewed_at ?? null,
      reviewer_note: row.reviewer_note ?? null,
    }
  })

  const pendingCount = requests.filter((r) => r.status === "pending").length

  return (
    <SidebarProvider>
      <AppSidebar
        user={{
          email: user.email ?? undefined,
          name: user.user_metadata?.full_name ?? user.user_metadata?.name,
        }}
        isAdmin
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
                <BreadcrumbPage>Admin</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-6 p-6">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Admin</h1>
            <p className="text-sm text-muted-foreground">
              Review and approve agent acquisition requests.
              {pendingCount > 0 && (
                <span className="ml-1 font-medium text-yellow-400">
                  {pendingCount} pending.
                </span>
              )}
            </p>
          </div>

          <AdminTable initialRequests={requests} />
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
