import { redirect } from "next/navigation";
import { createClient } from "@/lib/supabase/server";
import { providers } from "@/lib/providers";
import { ProviderCard } from "@/components/provider-card";
import { InstagramForm } from "@/components/instagram-form";

export default async function IntegrationsPage() {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user) {
    redirect("/login");
  }

  const { data: integrations } = await supabase
    .from("integrations")
    .select("id, provider, status, created_at")
    .eq("user_id", user.id);

  const integrationMap = new Map(
    (integrations ?? []).map((i) => [i.provider, i])
  );

  return (
    <div className="mx-auto max-w-5xl px-4 py-8">
      <h1 className="mb-2 text-2xl font-bold text-gray-900">Integrations</h1>
      <p className="mb-8 text-gray-500">
        Connect your accounts to enable AI agent access.
      </p>
      <div className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
        {providers.map((provider) => {
          const integration = integrationMap.get(provider.id) ?? null;
          return (
            <ProviderCard
              key={provider.id}
              provider={provider}
              integration={integration}
            >
              {provider.authType === "oauth" ? (
                <a
                  href={`/api/integrations/${provider.id}/connect`}
                  className="inline-block rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-800 transition-colors"
                >
                  Connect
                </a>
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
