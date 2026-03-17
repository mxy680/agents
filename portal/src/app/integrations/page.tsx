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

  if (!user) {
    redirect("/login");
  }

  // Get the integration catalog to map names → IDs
  const { data: catalog } = await supabase
    .from("integrations")
    .select("id, name");

  const catalogMap = new Map(
    (catalog ?? []).map((i) => [i.name, i.id])
  );

  // Get user's connected integrations (including account_label)
  const { data: userIntegrations } = await supabase
    .from("user_integrations")
    .select("id, integration_id, status, connected_at, account_label")
    .eq("user_id", user.id);

  // Group by integration_id so each provider can have multiple accounts
  const connectedMap = new Map<string, typeof userIntegrations>();
  for (const ui of userIntegrations ?? []) {
    const existing = connectedMap.get(ui.integration_id) ?? [];
    connectedMap.set(ui.integration_id, [...existing, ui]);
  }

  return (
    <div className="mx-auto max-w-4xl">
      <div className="mb-6">
        <h1 className="text-2xl font-bold tracking-tight">Integrations</h1>
        <p className="text-sm text-muted-foreground">
          Connect your accounts to enable AI agent access.
        </p>
      </div>
      {errorParam && (
        <Alert variant="destructive" className="mb-4">
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
            <ProviderCard
              key={provider.id}
              provider={provider}
              integrations={integrations}
            >
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
  );
}
