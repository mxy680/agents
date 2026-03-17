"use client";

import Link from "next/link";
import { buttonVariants } from "@/components/ui/button";

export default function Home() {
  return (
    <div className="flex min-h-[calc(100vh-6rem)] items-center justify-center">
      <div className="max-w-lg text-center">
        <h1 className="mb-3 text-3xl font-bold tracking-tight">
          Agent Marketplace
        </h1>
        <p className="mb-6 text-muted-foreground">
          Connect your Google, GitHub, and Instagram accounts so AI agents can
          work on your behalf. Tokens are encrypted at rest and never exposed
          directly.
        </p>
        <Link href="/integrations" className={buttonVariants()}>
          Get Started
        </Link>
      </div>
    </div>
  );
}
