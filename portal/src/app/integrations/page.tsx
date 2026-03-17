import { redirect } from "next/navigation";
import { createClient } from "@/lib/supabase/server";
import { providers } from "@/lib/providers";
import { ProviderCard } from "@/components/provider-card";
import { InstagramForm } from "@/components/instagram-form";
import { ConnectDialog } from "@/components/connect-dialog";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { AlertCircleIcon } from "lucide-react";

export default async function IntegrationsPage({
  searchParams,
}: {
  searchParams: Promise<{ error?: string }>;
}) {
  const { error: errorParam } = await searchParams;
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user) redirect("/login");

  const { data: catalog } = await supabase.from("integrations").select("id, name");
  const catalogMap = new Map((catalog ?? []).map((i) => [i.name, i.id]));

  const { data: userIntegrations } = await supabase
    .from("user_integrations")
    .select("id, integration_id, status, connected_at, account_label")
    .eq("user_id", user.id);

  const connectedMap = new Map<string, typeof userIntegrations>();
  for (const ui of userIntegrations ?? []) {
    const existing = connectedMap.get(ui.integration_id) ?? [];
    connectedMap.set(ui.integration_id, [...existing, ui]);
  }

  return (
    <div className="@container/main flex flex-1 flex-col gap-2">
      <div className="flex flex-col gap-4 px-4 py-6 lg:px-6">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Integrations</h1>
          <p className="text-sm text-muted-foreground">
            Connect your accounts to enable AI agent access.
          </p>
        </div>
        {errorParam && (
          <Alert variant="destructive">
            <AlertCircleIcon className="size-4" />
            <AlertDescription>{errorParam.replace(/_/g, " ")}</AlertDescription>
          </Alert>
        )}
        <div className="grid gap-4 sm:grid-cols-2">
          {providers.map((provider) => {
            const integrationId = catalogMap.get(provider.id);
            const rows = integrationId ? connectedMap.get(integrationId) ?? [] : [];
            const integrations = rows.map((ui) => ({
              id: ui.id,
              provider: provider.id,
              status: ui.status,
              account_label: ui.account_label ?? "",
              connected_at: ui.connected_at,
            }));
            return (
              <ProviderCard key={provider.id} provider={provider} integrations={integrations}>
                {provider.authType === "oauth" ? (
                  <ConnectDialog
                    href={`/api/integrations/${provider.id}/connect`}
                    providerName={provider.name}
                  />
                ) : provider.id === "instagram" ? (
                  <InstagramForm />
                ) : null}
              </ProviderCard>
            );
          })}
        </div>
      </div>
    </div>
  );
}
