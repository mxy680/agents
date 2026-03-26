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
import { ClientsTable } from "./clients-table"

export default async function ClientsPage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user) redirect("/login")
  if (!isAdmin(user.email)) redirect("/")

  const admin = createAdminClient()

  const { data: clients } = await admin
    .from("client_access")
    .select("id, code, client_name, agent_name, agent_names, active, created_at")
    .order("created_at", { ascending: false })

  const { data: templates } = await admin
    .from("agent_templates")
    .select("name, display_name")
    .eq("status", "active")

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
          <Separator orientation="vertical" className="mr-2 data-vertical:h-4 data-vertical:self-auto" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbPage>Clients</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-6 p-6">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Clients</h1>
            <p className="text-sm text-muted-foreground">
              Manage client access codes and agent assignments.
            </p>
          </div>
          <ClientsTable
            initialClients={clients ?? []}
            agents={(templates ?? []).map((t) => ({ name: t.name, displayName: t.display_name }))}
          />
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
