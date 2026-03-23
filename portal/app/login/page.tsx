import { LoginForm } from "@/components/login-form"
import { IconApps } from "@tabler/icons-react"

export default async function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ error?: string }>
}) {
  const { error } = await searchParams
  return (
    <div className="grid min-h-svh lg:grid-cols-2">
      {/* Left — branding panel */}
      <div className="relative hidden flex-col justify-between bg-foreground p-10 text-background lg:flex">
        <div className="flex items-center gap-3">
          <div className="flex size-8 items-center justify-center bg-primary">
            <IconApps className="size-5 text-primary-foreground" />
          </div>
          <span className="text-lg font-semibold tracking-tight">Emdash Admin</span>
        </div>

        <div className="flex flex-col gap-4">
          <p className="text-3xl font-semibold leading-snug tracking-tight">
            Manage integrations.<br />
            Deploy agents for clients.
          </p>
          <p className="text-sm text-background/60">
            Admin portal for managing integration credentials, configuring AI agents, and deploying them on behalf of clients.
          </p>
        </div>

        <div className="text-xs text-background/40">
          Internal tool — authorized access only
        </div>
      </div>

      {/* Right — form panel */}
      <div className="flex items-center justify-center p-8">
        <div className="flex w-full max-w-sm flex-col gap-8">
          {/* Mobile logo */}
          <div className="flex items-center gap-3 lg:hidden">
            <div className="flex size-8 items-center justify-center bg-primary">
              <IconApps className="size-5 text-primary-foreground" />
            </div>
            <span className="text-lg font-semibold tracking-tight">Emdash Admin</span>
          </div>

          <div className="flex flex-col gap-1">
            <h1 className="text-2xl font-semibold tracking-tight">Sign in</h1>
            <p className="text-sm text-muted-foreground">
              Choose a provider to continue
            </p>
          </div>

          {error && (
            <p className="text-center text-sm text-destructive">
              {error === "auth_callback_failed"
                ? "Authentication failed. Please try again."
                : decodeURIComponent(error)}
            </p>
          )}

          <LoginForm />
        </div>
      </div>
    </div>
  )
}
