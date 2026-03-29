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
import { ConnectDialog } from "@/components/connect-dialog"
import { PlaywrightConnectDialog } from "@/components/playwright-connect-dialog"
import { FramerConnectDialog } from "@/components/framer-connect-dialog"
import { BlueBubblesConnectDialog } from "@/components/bluebubbles-connect-dialog"
import { CookieCaptureButton } from "@/components/cookie-capture-button"
import { CookiePasteDialog } from "@/components/cookie-paste-dialog"
import { ApiKeyDialog } from "@/components/api-key-dialog"
import { GCPConnectDialog } from "@/components/gcp-connect-dialog"
import { AccountItem } from "@/components/account-item"
import {
  IconBrandGoogle,
  IconBrandGithub,
  IconBrandInstagram,
  IconBrandLinkedin,
  IconBrandX,
  IconLayout,
  IconBrandSupabase,
  IconMessage,
  IconSchool,
  IconPlus,
  IconBike,
  IconBuildingBank,
  IconChartLine,
  IconCoffin,
  IconBottle,
  IconUsers,
  IconBuildingCommunity,
  IconMapPin,
  IconCertificate,
  IconCheck,
  IconBrandVercel,
  IconBrandCloudflare,
  IconLayoutKanban,
  IconRocket,
  IconCloud,
} from "@tabler/icons-react"

const providers: Array<{ id: string; name: string; description: string; icon: typeof IconCheck; connectType: string }> = [
  {
    id: "google",
    name: "Google",
    description: "Gmail, Sheets, Calendar, Drive",
    icon: IconBrandGoogle,
    connectType: "oauth" as const,
  },
  {
    id: "github",
    name: "GitHub",
    description: "Repos, Issues, Pull Requests, Actions",
    icon: IconBrandGithub,
    connectType: "oauth" as const,
  },
  {
    id: "instagram",
    name: "Instagram",
    description: "Media, Stories, Comments, Messages",
    icon: IconBrandInstagram,
    connectType: "cookie-paste" as const,
  },
  {
    id: "linkedin",
    name: "LinkedIn",
    description: "Posts, Connections, Messages, Jobs",
    icon: IconBrandLinkedin,
    connectType: "cookie-paste" as const,
  },
  {
    id: "x",
    name: "X",
    description: "Posts, Likes, DMs, Lists, Communities",
    icon: IconBrandX,
    connectType: "cookie-paste" as const,
  },
  {
    id: "framer",
    name: "Framer",
    description: "Pages, Collections, Styles, Deployments",
    icon: IconLayout,
    connectType: "framer" as const,
  },
  {
    id: "supabase",
    name: "Supabase",
    description: "Projects, Branches, Auth, Database",
    icon: IconBrandSupabase,
    connectType: "oauth" as const,
  },
  {
    id: "bluebubbles",
    name: "iMessage",
    description: "Chats, Messages, Attachments, FaceTime",
    icon: IconMessage,
    connectType: "bluebubbles" as const,
  },
  {
    id: "canvas",
    name: "Canvas LMS",
    description: "Courses, Assignments, Grades, Discussions",
    icon: IconSchool,
    connectType: "cookie-paste" as const,
  },
  {
    id: "vercel",
    name: "Vercel",
    description: "Projects, Deployments, Domains",
    icon: IconBrandVercel,
    connectType: "api-key" as const,
  },
  {
    id: "cloudflare",
    name: "Cloudflare",
    description: "Zones, DNS, Workers, Pages, R2, KV",
    icon: IconBrandCloudflare,
    connectType: "api-key" as const,
  },
  {
    id: "linear",
    name: "Linear",
    description: "Issues, Projects, Cycles, Teams",
    icon: IconLayoutKanban,
    connectType: "api-key" as const,
  },
  {
    id: "fly",
    name: "Fly.io",
    description: "Apps, Machines, Volumes, Certificates, Secrets",
    icon: IconRocket,
    connectType: "api-key" as const,
  },
  {
    id: "gcp",
    name: "Google Cloud Platform",
    description: "Projects, APIs, OAuth clients, IAM, Service Accounts",
    icon: IconCloud,
    connectType: "gcp" as const,
  },
  {
    id: "citibike",
    name: "Citi Bike",
    description: "Station density, transit accessibility signals",
    icon: IconBike,
    connectType: "none" as const,
  },
  {
    id: "hmda",
    name: "HMDA",
    description: "Mortgage originations, investor activity by census tract",
    icon: IconBuildingBank,
    connectType: "none" as const,
  },
  {
    id: "census",
    name: "Census ACS",
    description: "Demographics, income, rent burden, vacancy by tract",
    icon: IconUsers,
    connectType: "none" as const,
  },
  {
    id: "nydos",
    name: "NY Dept of State",
    description: "LLC formations, entity search, address matching",
    icon: IconCertificate,
    connectType: "none" as const,
  },
  {
    id: "dof",
    name: "NYC Dept of Finance",
    description: "Property owners, tax assessments, entity lookup",
    icon: IconBuildingCommunity,
    connectType: "none" as const,
  },
  {
    id: "obituaries",
    name: "Obituaries",
    description: "Estate property detection via Legacy.com",
    icon: IconCoffin,
    connectType: "none" as const,
  },
  {
    id: "trends",
    name: "Google Trends",
    description: "Neighborhood search momentum signals",
    icon: IconChartLine,
    connectType: "none" as const,
  },
  {
    id: "places",
    name: "Google Places",
    description: "Nearby amenities, transit, business activity",
    icon: IconMapPin,
    connectType: "none" as const,
  },
  {
    id: "nyscef",
    name: "NYSCEF",
    description: "Court records, foreclosure proceedings",
    icon: IconBuildingBank,
    connectType: "none" as const,
  },
]

export default async function IntegrationsPage() {
  const supabase = await createClient()
  const { data: { user } } = await supabase.auth.getUser()

  if (!user) {
    redirect("/login")
  }

  if (!isAdmin(user.email)) {
    redirect("/login?error=not_authorized")
  }

  // Fetch all active integrations using admin client (no RLS)
  const admin = createAdminClient()
  const { data: integrations } = await admin
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
                <BreadcrumbPage>Integrations</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
        </header>
        <div className="flex flex-1 flex-col gap-6 p-6">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Integrations</h1>
            <p className="text-sm text-muted-foreground">
              Manage centralized integration credentials for all agents.
            </p>
          </div>
          <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {providers.map((provider) => {
              const accounts = integrationsByProvider[provider.id] ?? []
              return (
                <Card key={provider.id} id={provider.id} className="flex flex-col">
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
                  <CardContent className="flex flex-col gap-3 mt-auto">
                    {accounts.length > 0 && (
                      <div className="flex flex-col gap-2">
                        {accounts.map((account) => (
                          <AccountItem
                            key={account.id}
                            id={account.id}
                            label={account.label}
                            status={account.status}
                            provider={provider.id}
                            connectType={provider.connectType}
                          />
                        ))}
                      </div>
                    )}
                    {provider.connectType === "none" ? (
                      <div className="flex items-center gap-1.5 text-xs text-green-500">
                        <IconCheck className="size-3.5" />
                        <span>No auth required — public API</span>
                      </div>
                    ) : provider.connectType === "playwright" ? (
                      <PlaywrightConnectDialog
                        provider={provider.id}
                        providerName={provider.name}
                      >
                        <Button variant="outline" size="sm" className="w-full">
                          <IconPlus />
                          {accounts.length > 0 ? "Add account" : "Launch browser"}
                        </Button>
                      </PlaywrightConnectDialog>
                    ) : provider.connectType === "framer" ? (
                      <FramerConnectDialog>
                        <Button variant="outline" size="sm" className="w-full">
                          <IconPlus />
                          {accounts.length > 0 ? "Add project" : "Connect"}
                        </Button>
                      </FramerConnectDialog>
                    ) : provider.connectType === "bluebubbles" ? (
                      <BlueBubblesConnectDialog>
                        <Button variant="outline" size="sm" className="w-full">
                          <IconPlus />
                          {accounts.length > 0 ? "Add server" : "Connect"}
                        </Button>
                      </BlueBubblesConnectDialog>
                    ) : provider.connectType === "cookie-capture" ? (
                      <CookieCaptureButton provider={provider.id}>
                        <Button variant="outline" size="sm" className="w-full">
                          <IconPlus />
                          {accounts.length > 0 ? "Refresh cookies" : "Capture cookies"}
                        </Button>
                      </CookieCaptureButton>
                    ) : provider.connectType === "gcp" ? (
                      <GCPConnectDialog>
                        <Button variant="outline" size="sm" className="w-full">
                          <IconPlus />
                          {accounts.length > 0 ? "Update token" : "Connect"}
                        </Button>
                      </GCPConnectDialog>
                    ) : provider.connectType === "api-key" ? (
                      <ApiKeyDialog
                        provider={provider.id}
                        providerName={provider.name}
                      >
                        <Button variant="outline" size="sm" className="w-full">
                          <IconPlus />
                          {accounts.length > 0 ? "Update token" : "Connect"}
                        </Button>
                      </ApiKeyDialog>
                    ) : provider.connectType === "cookie-paste" ? (
                      <CookiePasteDialog
                        provider={provider.id}
                        providerName={provider.name}
                      >
                        <Button variant="outline" size="sm" className="w-full">
                          <IconPlus />
                          {accounts.length > 0 ? "Refresh cookies" : "Paste cookies"}
                        </Button>
                      </CookiePasteDialog>
                    ) : (
                      <ConnectDialog
                        provider={provider.id}
                        providerName={provider.name}
                      >
                        <Button variant="outline" size="sm" className="w-full">
                          <IconPlus />
                          {accounts.length > 0 ? "Add account" : "Connect"}
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
