"use client";

import { useRouter } from "next/navigation";
import type { ProviderMeta } from "@/lib/providers";

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
    <div className="rounded-lg border border-gray-200 bg-white p-6 shadow-sm">
      <div className="mb-4 flex items-center justify-between">
        <h3 className="text-lg font-semibold text-gray-900">{provider.name}</h3>
        <span
          className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium ${
            connected
              ? "bg-green-100 text-green-800"
              : "bg-gray-100 text-gray-600"
          }`}
        >
          {connected ? "Connected" : "Not connected"}
        </span>
      </div>
      <p className="mb-4 text-sm text-gray-500">{provider.description}</p>
      {connected ? (
        <button
          onClick={handleDisconnect}
          className="rounded-md border border-red-300 px-4 py-2 text-sm font-medium text-red-700 hover:bg-red-50 transition-colors"
        >
          Disconnect
        </button>
      ) : (
        children
      )}
    </div>
  );
}
