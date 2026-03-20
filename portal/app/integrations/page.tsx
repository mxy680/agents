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
import { ConnectDialog } from "@/components/connect-dialog"
import { InstagramConnectDialog } from "@/components/instagram-connect-dialog"
import { LinkedinConnectDialog } from "@/components/linkedin-connect-dialog"
import { FramerConnectDialog } from "@/components/framer-connect-dialog"
import { AccountItem } from "@/components/account-item"
import {
  IconBrandGoogle,
  IconBrandGithub,
  IconBrandInstagram,
  IconBrandLinkedin,
  IconLayout,
  IconBrandSupabase,
  IconPlus,
} from "@tabler/icons-react"

const providers = [
  {
    id: "google",
    name: "Google",
    description: "Gmail, Sheets, Calendar, Drive",
    icon: IconBrandGoogle,
  },
  {
    id: "github",
    name: "GitHub",
    description: "Repos, Issues, Pull Requests, Actions",
    icon: IconBrandGithub,
  },
  {
    id: "instagram",
    name: "Instagram",
    description: "Media, Stories, Comments, Messages",
    icon: IconBrandInstagram,
  },
  {
    id: "linkedin",
    name: "LinkedIn",
    description: "Posts, Connections, Messages, Jobs",
    icon: IconBrandLinkedin,
  },
  {
    id: "framer",
    name: "Framer",
    description: "Pages, Collections, Styles, Deployments",
    icon: IconLayout,
  },
  {
    id: "supabase",
    name: "Supabase",
    description: "Projects, Branches, Auth, Database",
    icon: IconBrandSupabase,
  },
]

export default async function IntegrationsPage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user) {
    redirect("/login")
  }

  // Fetch user's connected integrations
  const { data: integrations } = await supabase
    .from("user_integrations")
    .select("id, provider, label, status, created_at")
    .eq("user_id", user.id)
    .order("created_at", { ascending: true })

  const integrationsByProvider = (integrations ?? []).reduce<
    Record<string, typeof integrations>
  >((acc, integration) => {
    const key = integration.provider
    if (!acc[key]) acc[key] = []
    acc[key]!.push(integration)
    return acc
  }, {})

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
                <BreadcrumbPage>Integrations</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-6 p-6">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Integrations</h1>
            <p className="text-sm text-muted-foreground">
              Connect your accounts to let agents access external services.
            </p>
          </div>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {providers.map((provider) => {
              const accounts = integrationsByProvider[provider.id] ?? []
              return (
                <Card key={provider.id} id={provider.id}>
                  <CardHeader>
                    <div className="flex items-center gap-3">
                      <div className="flex size-10 items-center justify-center bg-muted">
                        <provider.icon className="size-5" />
                      </div>
                      <div className="flex-1">
                        <CardTitle className="text-base">{provider.name}</CardTitle>
                        <CardDescription>{provider.description}</CardDescription>
                      </div>
                    </div>
                  </CardHeader>
                  <CardContent className="flex flex-col gap-3">
                    {accounts.length > 0 && (
                      <div className="flex flex-col gap-2">
                        {accounts.map((account) => (
                          <AccountItem
                            key={account.id}
                            id={account.id}
                            label={account.label}
                            status={account.status}
                          />
                        ))}
                      </div>
                    )}
                    {provider.id === "instagram" ? (
                      <InstagramConnectDialog>
                        <Button variant="outline" size="sm" className="w-full">
                          <IconPlus />
                          {accounts.length > 0
                            ? "Add another account"
                            : "Connect"}
                        </Button>
                      </InstagramConnectDialog>
                    ) : provider.id === "linkedin" ? (
                      <LinkedinConnectDialog>
                        <Button variant="outline" size="sm" className="w-full">
                          <IconPlus />
                          {accounts.length > 0
                            ? "Add another account"
                            : "Connect"}
                        </Button>
                      </LinkedinConnectDialog>
                    ) : provider.id === "framer" ? (
                      <FramerConnectDialog>
                        <Button variant="outline" size="sm" className="w-full">
                          <IconPlus />
                          {accounts.length > 0
                            ? "Add another account"
                            : "Connect"}
                        </Button>
                      </FramerConnectDialog>
                    ) : (
                      <ConnectDialog
                        provider={provider.id}
                        providerName={provider.name}
                      >
                        <Button variant="outline" size="sm" className="w-full">
                          <IconPlus />
                          {accounts.length > 0
                            ? "Add another account"
                            : "Connect"}
                        </Button>
                      </ConnectDialog>
                    )}
                  </CardContent>
                </Card>
              )
            })}
          </div>
        </div>
      </SidebarInset>
    </SidebarProvider>
  )
}
