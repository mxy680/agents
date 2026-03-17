"use client";

import { useRouter } from "next/navigation";
import { useState, useRef } from "react";
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
import { Input } from "@/components/ui/input";
import { MailIcon, GithubIcon, InstagramIcon, PlugIcon } from "lucide-react";

const PROVIDER_ICONS: Record<string, React.ComponentType<{ className?: string }>> = {
  google: MailIcon,
  github: GithubIcon,
  instagram: InstagramIcon,
};

interface Integration {
  id: string;
  provider: string;
  status: string;
  account_label: string;
  connected_at: string;
}

function AccountRow({
  integ,
  providerId,
  onRefresh,
}: {
  integ: Integration;
  providerId: string;
  onRefresh: () => void;
}) {
  const [editing, setEditing] = useState(false);
  const [value, setValue] = useState(integ.account_label);
  const [saving, setSaving] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);

  async function handleRename() {
    const newLabel = value.trim();
    if (newLabel === integ.account_label) { setEditing(false); return; }
    setSaving(true);
    await fetch(`/api/integrations/${providerId}/rename`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ old_label: integ.account_label, new_label: newLabel }),
    });
    setSaving(false);
    setEditing(false);
    onRefresh();
  }

  async function handleDisconnect() {
    await fetch(`/api/integrations/${providerId}/disconnect`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ account_label: integ.account_label }),
    });
    onRefresh();
  }

  if (editing) {
    return (
      <li className="flex items-center gap-2 rounded-md border px-3 py-2 text-sm">
        <Input
          ref={inputRef}
          className="h-6 flex-1 border-0 p-0 text-sm shadow-none focus-visible:ring-0"
          value={value}
          onChange={(e) => setValue(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") handleRename();
            if (e.key === "Escape") { setValue(integ.account_label); setEditing(false); }
          }}
          onBlur={handleRename}
          disabled={saving}
          autoFocus
        />
        <Button variant="ghost" size="sm" className="shrink-0 text-xs" onClick={handleRename} disabled={saving}>
          {saving ? "…" : "Save"}
        </Button>
      </li>
    );
  }

  return (
    <li className="flex items-center justify-between rounded-md border px-3 py-2 text-sm">
      <button
        className="truncate text-muted-foreground hover:text-foreground text-left"
        onClick={() => setEditing(true)}
        title="Click to rename"
      >
        {integ.account_label || "(no label)"}
      </button>
      <Button
        variant="ghost"
        size="sm"
        className="ml-2 shrink-0 text-destructive hover:text-destructive"
        onClick={handleDisconnect}
      >
        Disconnect
      </Button>
    </li>
  );
}

export function ProviderCard({
  provider,
  integrations,
  children,
}: {
  provider: ProviderMeta;
  integrations: Integration[];
  children?: React.ReactNode;
}) {
  const router = useRouter();
  const connectedCount = integrations.filter((i) => i.status === "connected").length;
  const Icon = PROVIDER_ICONS[provider.id] ?? PlugIcon;

  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between space-y-0 pb-3">
        <div className="flex items-center gap-3">
          <div className="flex size-9 items-center justify-center rounded-lg bg-muted">
            <Icon className="size-4 text-muted-foreground" />
          </div>
          <CardTitle className="text-base font-semibold">{provider.name}</CardTitle>
        </div>
        <Badge variant={connectedCount > 0 ? "default" : "secondary"} className="shrink-0">
          {connectedCount > 0 ? `${connectedCount} connected` : "Not connected"}
        </Badge>
      </CardHeader>
      <CardContent>
        <CardDescription className="mb-4">{provider.description}</CardDescription>
        {integrations.length > 0 && (
          <ul className="mb-4 space-y-2">
            {integrations.map((integ) => (
              <AccountRow
                key={integ.id}
                integ={integ}
                providerId={provider.id}
                onRefresh={() => router.refresh()}
              />
            ))}
          </ul>
        )}
        {children}
      </CardContent>
    </Card>
  );
}
