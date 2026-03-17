import Link from "next/link";
import { KeyRoundIcon, ZapIcon, ShieldCheckIcon, TerminalIcon } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader } from "@/components/ui/card";

const FEATURES = [
  {
    icon: ZapIcon,
    title: "One-click OAuth",
    description: "Connect Google, GitHub, and more with secure OAuth flows.",
  },
  {
    icon: ShieldCheckIcon,
    title: "Encrypted at rest",
    description: "Tokens are encrypted with AES-256-GCM before storage.",
  },
  {
    icon: TerminalIcon,
    title: "Agent-ready",
    description: "CLI bridge exports tokens as env vars for AI agent containers.",
  },
];

export default function HomePage() {
  return (
    <div className="@container/main flex flex-1 flex-col gap-2">
      <div className="flex flex-col items-center justify-center gap-10 px-4 py-16 md:py-24 text-center">
        <div className="flex flex-col items-center gap-6 max-w-2xl">
          <div className="flex size-16 items-center justify-center rounded-2xl bg-muted">
            <KeyRoundIcon className="size-8 text-muted-foreground" />
          </div>
          <div className="space-y-3">
            <h1 className="text-4xl font-bold tracking-tight md:text-5xl">Agent Marketplace</h1>
            <p className="text-lg text-muted-foreground">
              Connect your accounts so AI agents can work on your behalf.
              Tokens are encrypted at rest and never exposed directly.
            </p>
          </div>
          <div className="flex items-center gap-3">
            <Button asChild size="lg">
              <Link href="/integrations">Get Started</Link>
            </Button>
            <Button asChild variant="outline" size="lg">
              <Link href="/login">Sign In</Link>
            </Button>
          </div>
        </div>

        <div className="grid gap-4 sm:grid-cols-3 w-full max-w-3xl">
          {FEATURES.map((f) => (
            <Card key={f.title}>
              <CardHeader className="pb-2">
                <div className="flex size-8 items-center justify-center rounded-lg bg-muted mb-2">
                  <f.icon className="size-4 text-muted-foreground" />
                </div>
                <p className="font-medium text-sm">{f.title}</p>
              </CardHeader>
              <CardContent>
                <CardDescription>{f.description}</CardDescription>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </div>
  );
}
