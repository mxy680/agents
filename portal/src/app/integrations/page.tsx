import { redirect } from "next/navigation";
import { createClient } from "@/lib/supabase/server";
import { providers } from "@/lib/providers";
import { ProviderCard } from "@/components/provider-card";
import { InstagramForm } from "@/components/instagram-form";
import { ConnectDialog } from "@/components/connect-dialog";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Separator } from "@/components/ui/separator";
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
    <div className="@container/main flex flex-1 flex-col">
      <div className="mx-auto w-full max-w-2xl px-4 py-8 lg:px-6">
        <div className="mb-6">
          <h1 className="text-xl font-semibold tracking-tight">Integrations</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Connect accounts to give AI agents access to your services.
          </p>
        </div>

        {errorParam && (
          <Alert variant="destructive" className="mb-6">
            <AlertCircleIcon className="size-4" />
            <AlertDescription>{errorParam.replace(/_/g, " ")}</AlertDescription>
          </Alert>
        )}

        <div className="rounded-xl border bg-card">
          {providers.map((provider, i) => {
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
              <div key={provider.id}>
                {i > 0 && <Separator />}
                <div className="px-5">
                  <ProviderCard provider={provider} integrations={integrations}>
                    {provider.authType === "oauth" ? (
                      <ConnectDialog
                        href={`/api/integrations/${provider.id}/connect`}
                        providerName={provider.name}
                      />
                    ) : provider.id === "instagram" ? (
                      <InstagramForm />
                    ) : null}
                  </ProviderCard>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
