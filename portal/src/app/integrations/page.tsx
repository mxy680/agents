import { redirect } from "next/navigation";
import { createClient } from "@/lib/supabase/server";
import { providers } from "@/lib/providers";
import { ProviderCard } from "@/components/provider-card";
import { InstagramForm } from "@/components/instagram-form";
import { ConnectButton } from "@/components/connect-button";

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
        <div className="mb-4 rounded-md border border-destructive/50 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          Error: {errorParam.replace(/_/g, " ")}
        </div>
      )}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
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
                <ConnectButton
                  href={`/api/integrations/${provider.id}/connect`}
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
