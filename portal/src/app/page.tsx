"use client";

import Link from "next/link";
import { PlugZapIcon, ShieldCheckIcon, KeyRoundIcon, ZapIcon } from "lucide-react";
import { buttonVariants } from "@/components/ui/button";

const FEATURES = [
  {
    icon: PlugZapIcon,
    title: "One-click OAuth",
    description: "Connect Google, GitHub, and more with secure OAuth flows.",
  },
  {
    icon: ShieldCheckIcon,
    title: "Encrypted at rest",
    description: "Tokens are encrypted with AES-256-GCM before storage.",
  },
  {
    icon: ZapIcon,
    title: "Agent-ready",
    description: "CLI bridge exports tokens as env vars for AI agent containers.",
  },
];

export default function Home() {
  return (
    <div className="flex min-h-[calc(100vh-6rem)] flex-col items-center justify-center gap-12 py-12">
      <div className="flex max-w-xl flex-col items-center gap-6 text-center">
        <div className="flex size-16 items-center justify-center rounded-2xl bg-primary/10 ring-1 ring-primary/20">
          <KeyRoundIcon className="size-8 text-primary" />
        </div>
        <div className="space-y-3">
          <h1 className="text-4xl font-bold tracking-tight">Agent Marketplace</h1>
          <p className="text-lg text-muted-foreground">
            Connect your accounts so AI agents can work on your behalf. Tokens
            are encrypted at rest and never exposed directly.
          </p>
        </div>
        <div className="flex items-center gap-3">
          <Link href="/integrations" className={buttonVariants({ size: "lg" })}>
            Get Started
          </Link>
          <Link
            href="/login"
            className={buttonVariants({ variant: "outline", size: "lg" })}
          >
            Sign In
          </Link>
        </div>
      </div>

      <div className="grid w-full max-w-2xl gap-4 sm:grid-cols-3">
        {FEATURES.map((f) => (
          <div
            key={f.title}
            className="flex flex-col gap-2 rounded-xl border bg-card p-4 text-card-foreground"
          >
            <div className="flex size-8 items-center justify-center rounded-lg bg-muted">
              <f.icon className="size-4 text-muted-foreground" />
            </div>
            <p className="text-sm font-medium">{f.title}</p>
            <p className="text-xs text-muted-foreground">{f.description}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
