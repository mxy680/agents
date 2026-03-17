import { redirect } from "next/navigation";
import { createClient } from "@/lib/supabase/server";
import { providers } from "@/lib/providers";
import { ProviderCard } from "@/components/provider-card";
import { InstagramForm } from "@/components/instagram-form";
import { ConnectButton } from "@/components/connect-button";

export default async function IntegrationsPage() {
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

  // Get user's connected integrations
  const { data: userIntegrations } = await supabase
    .from("user_integrations")
    .select("id, integration_id, status, connected_at")
    .eq("user_id", user.id);

  // Map integration_id → user_integration row
  const connectedMap = new Map(
    (userIntegrations ?? []).map((ui) => [ui.integration_id, ui])
  );

  return (
    <div className="mx-auto max-w-4xl">
      <div className="mb-6">
        <h1 className="text-2xl font-bold tracking-tight">Integrations</h1>
        <p className="text-sm text-muted-foreground">
          Connect your accounts to enable AI agent access.
        </p>
      </div>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {providers.map((provider) => {
          const integrationId = catalogMap.get(provider.id);
          const userInteg = integrationId
            ? connectedMap.get(integrationId) ?? null
            : null;
          return (
            <ProviderCard
              key={provider.id}
              provider={provider}
              integration={
                userInteg
                  ? {
                      id: userInteg.id,
                      provider: provider.id,
                      status: userInteg.status,
                      created_at: userInteg.connected_at,
                    }
                  : null
              }
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
