"use client";

import { useRouter } from "next/navigation";
import { useState, useRef } from "react";
import type { ProviderMeta } from "@/lib/providers";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
  MailIcon,
  GithubIcon,
  InstagramIcon,
  PlugIcon,
  CheckCircle2Icon,
  PencilIcon,
  XIcon,
} from "lucide-react";
import { cn } from "@/lib/utils";

const PROVIDER_ICONS: Record<string, React.ComponentType<{ className?: string }>> = {
  google: MailIcon,
  github: GithubIcon,
  instagram: InstagramIcon,
};

const PROVIDER_COLORS: Record<string, string> = {
  google: "bg-red-50 text-red-600 dark:bg-red-950 dark:text-red-400",
  github: "bg-zinc-100 text-zinc-700 dark:bg-zinc-800 dark:text-zinc-300",
  instagram: "bg-purple-50 text-purple-600 dark:bg-purple-950 dark:text-purple-400",
};

interface Integration {
  id: string;
  provider: string;
  status: string;
  account_label: string;
  connected_at: string;
}

function AccountPill({
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
      <span className="inline-flex items-center gap-1 rounded-full border bg-background pl-3 pr-1 py-1 text-xs">
        <Input
          className="h-4 w-24 border-0 p-0 text-xs shadow-none focus-visible:ring-0"
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
        <button
          className="rounded-full p-0.5 text-muted-foreground hover:text-foreground"
          onClick={() => { setValue(integ.account_label); setEditing(false); }}
        >
          <XIcon className="size-3" />
        </button>
      </span>
    );
  }

  return (
    <span className="group inline-flex items-center gap-1.5 rounded-full border bg-background pl-2.5 pr-1 py-1 text-xs">
      <CheckCircle2Icon className="size-3 text-emerald-500 shrink-0" />
      <span className="text-foreground">{integ.account_label || "(no label)"}</span>
      <button
        className="rounded-full p-0.5 text-muted-foreground opacity-0 group-hover:opacity-100 hover:text-foreground transition-opacity"
        onClick={() => setEditing(true)}
        title="Rename"
      >
        <PencilIcon className="size-3" />
      </button>
      <button
        className="rounded-full p-0.5 text-muted-foreground opacity-0 group-hover:opacity-100 hover:text-destructive transition-opacity"
        onClick={handleDisconnect}
        title="Disconnect"
      >
        <XIcon className="size-3" />
      </button>
    </span>
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
  const iconColor = PROVIDER_COLORS[provider.id] ?? "bg-muted text-muted-foreground";

  return (
    <div className="flex items-start gap-4 py-5">
      <div className={cn("flex size-10 shrink-0 items-center justify-center rounded-xl", iconColor)}>
        <Icon className="size-5" />
      </div>

      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 mb-0.5">
          <span className="font-medium text-sm">{provider.name}</span>
          {connectedCount > 0 && (
            <Badge variant="secondary" className="text-xs px-1.5 py-0 h-4 font-normal">
              {connectedCount} connected
            </Badge>
          )}
        </div>
        <p className="text-xs text-muted-foreground mb-3">{provider.description}</p>

        {integrations.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-3">
            {integrations.map((integ) => (
              <AccountPill
                key={integ.id}
                integ={integ}
                providerId={provider.id}
                onRefresh={() => router.refresh()}
              />
            ))}
          </div>
        )}

        {children}
      </div>
    </div>
  );
}
