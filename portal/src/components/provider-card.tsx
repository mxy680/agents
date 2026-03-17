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
  created_at: string;
}

interface ProviderCardProps {
  provider: ProviderMeta;
  integration: Integration | null;
  children?: React.ReactNode;
}

export function ProviderCard({
  provider,
  integration,
  children,
}: ProviderCardProps) {
  const router = useRouter();
  const connected = integration?.status === "active";

  async function handleDisconnect() {
    const res = await fetch(`/api/integrations/${provider.id}/disconnect`, {
      method: "POST",
    });
    if (res.ok) {
      router.refresh();
    }
  }

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
        <CardTitle className="text-base font-medium">{provider.name}</CardTitle>
        <Badge variant={connected ? "default" : "secondary"}>
          {connected ? "Connected" : "Not connected"}
        </Badge>
      </CardHeader>
      <CardContent>
        <CardDescription className="mb-4">
          {provider.description}
        </CardDescription>
        {connected ? (
          <Button variant="outline" size="sm" onClick={handleDisconnect}>
            Disconnect
          </Button>
        ) : (
          children
        )}
      </CardContent>
    </Card>
  );
}
