"use client";

import { useRouter } from "next/navigation";
import type { ProviderMeta } from "@/lib/providers";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";

interface Integration {
  id: string;
  provider: string;
  status: string;
  account_label: string;
  connected_at: string;
}

interface ProviderCardProps {
  provider: ProviderMeta;
  integrations: Integration[];
  children?: React.ReactNode;
}

export function ProviderCard({
  provider,
  integrations,
  children,
}: ProviderCardProps) {
  const router = useRouter();
  const connectedCount = integrations.filter((i) => i.status === "connected").length;

  async function handleDisconnect(accountLabel: string) {
    await fetch(`/api/integrations/${provider.id}/disconnect`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ account_label: accountLabel }),
    });
    router.refresh();
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-base font-medium">{provider.name}</CardTitle>
        <Badge variant={connectedCount > 0 ? "default" : "secondary"}>
          {connectedCount > 0
            ? `${connectedCount} connected`
            : "Not connected"}
        </Badge>
      </CardHeader>
      <CardContent>
        <CardDescription className="mb-4">
          {provider.description}
        </CardDescription>
        {integrations.length > 0 && (
          <ul className="mb-4 space-y-2">
            {integrations.map((integ) => (
              <li
                key={integ.id}
                className="flex items-center justify-between rounded-md border px-3 py-2 text-sm"
              >
                <span className="truncate text-muted-foreground">
                  {integ.account_label || "(no label)"}
                </span>
                <Button
                  variant="ghost"
                  size="sm"
                  className="ml-2 shrink-0 text-destructive hover:text-destructive"
                  onClick={() => handleDisconnect(integ.account_label)}
                >
                  Disconnect
                </Button>
              </li>
            ))}
          </ul>
        )}
        {children}
      </CardContent>
    </Card>
  );
}
