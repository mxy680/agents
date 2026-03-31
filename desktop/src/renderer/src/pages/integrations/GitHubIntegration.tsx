import { useCallback, useEffect, useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  GitBranch,
  RefreshCw,
  CheckCircle2,
  XCircle,
  Loader2,
  User,
  Shield
} from 'lucide-react'

interface GitDetectionResult {
  readonly installed: boolean
  readonly path: string | null
  readonly version: string | null
  readonly error: string | null
}

interface GhAuthResult {
  readonly authenticated: boolean
  readonly username: string | null
  readonly scopes: readonly string[]
  readonly error: string | null
}

interface GitHubDetectionResult {
  readonly git: GitDetectionResult
  readonly gh: GitDetectionResult
  readonly auth: GhAuthResult
}

export function GitHubIntegration(): React.JSX.Element {
  const [result, setResult] = useState<GitHubDetectionResult | null>(null)
  const [detecting, setDetecting] = useState(false)

  const detect = useCallback(async () => {
    setDetecting(true)
    try {
      const detection = await window.api.github.detect()
      setResult(detection)
    } catch (err) {
      setResult({
        git: { installed: false, path: null, version: null, error: err instanceof Error ? err.message : 'Detection failed' },
        gh: { installed: false, path: null, version: null, error: null },
        auth: { authenticated: false, username: null, scopes: [], error: null }
      })
    } finally {
      setDetecting(false)
    }
  }, [])

  useEffect(() => {
    detect()
  }, [detect])

  const overallConnected = result?.git.installed && result?.gh.installed && result?.auth.authenticated

  return (
    <div className="mx-auto max-w-2xl space-y-8 p-8">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-4">
          <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-secondary">
            <GitBranch className="h-6 w-6" />
          </div>
          <div>
            <h1 className="text-xl font-semibold">GitHub</h1>
            <p className="text-sm text-muted-foreground">
              Git version control and GitHub CLI
            </p>
          </div>
        </div>
        {result && (
          <Badge variant={overallConnected ? 'success' : 'destructive'}>
            {overallConnected ? 'Connected' : 'Incomplete'}
          </Badge>
        )}
      </div>

      {/* Git CLI */}
      <div className="rounded-lg border bg-card p-6">
        <h2 className="mb-4 text-sm font-medium">Git</h2>

        {detecting ? (
          <LoadingRow label="Detecting git..." />
        ) : result ? (
          <div className="space-y-3">
            <StatusRow
              label="Status"
              value={<InstalledBadge installed={result.git.installed} />}
            />
            {result.git.path && (
              <StatusRow label="Path" value={<code className="text-xs">{result.git.path}</code>} />
            )}
            {result.git.version && <StatusRow label="Version" value={result.git.version} />}
            {result.git.error && (
              <StatusRow label="Error" value={<span className="text-destructive">{result.git.error}</span>} />
            )}
          </div>
        ) : null}
      </div>

      {/* GitHub CLI */}
      <div className="rounded-lg border bg-card p-6">
        <h2 className="mb-4 text-sm font-medium">GitHub CLI (gh)</h2>

        {detecting ? (
          <LoadingRow label="Detecting gh..." />
        ) : result ? (
          <div className="space-y-3">
            <StatusRow
              label="Status"
              value={<InstalledBadge installed={result.gh.installed} />}
            />
            {result.gh.path && (
              <StatusRow label="Path" value={<code className="text-xs">{result.gh.path}</code>} />
            )}
            {result.gh.version && <StatusRow label="Version" value={result.gh.version} />}
            {result.gh.error && (
              <StatusRow label="Error" value={<span className="text-destructive">{result.gh.error}</span>} />
            )}
            {!result.gh.installed && (
              <p className="text-xs text-muted-foreground pt-1">
                Install with{' '}
                <code className="rounded bg-muted px-1 py-0.5">brew install gh</code>
              </p>
            )}
          </div>
        ) : null}
      </div>

      {/* Auth Status */}
      <div className="rounded-lg border bg-card p-6">
        <h2 className="mb-4 text-sm font-medium">Authentication</h2>

        {detecting ? (
          <LoadingRow label="Checking auth..." />
        ) : result ? (
          <div className="space-y-3">
            <StatusRow
              label="Status"
              value={
                result.auth.authenticated ? (
                  <span className="flex items-center gap-1.5 text-emerald-500">
                    <CheckCircle2 className="h-4 w-4" />
                    Authenticated
                  </span>
                ) : (
                  <span className="flex items-center gap-1.5 text-destructive">
                    <XCircle className="h-4 w-4" />
                    Not authenticated
                  </span>
                )
              }
            />
            {result.auth.username && (
              <StatusRow
                label="Account"
                value={
                  <span className="flex items-center gap-1.5">
                    <User className="h-3.5 w-3.5 text-muted-foreground" />
                    {result.auth.username}
                  </span>
                }
              />
            )}
            {result.auth.scopes.length > 0 && (
              <StatusRow
                label="Scopes"
                value={
                  <span className="flex items-center gap-1.5">
                    <Shield className="h-3.5 w-3.5 text-muted-foreground" />
                    <span className="text-xs">{result.auth.scopes.join(', ')}</span>
                  </span>
                }
              />
            )}
            {result.auth.error && (
              <StatusRow label="Error" value={<span className="text-destructive">{result.auth.error}</span>} />
            )}
            {!result.auth.authenticated && result.gh.installed && (
              <p className="text-xs text-muted-foreground pt-1">
                Run{' '}
                <code className="rounded bg-muted px-1 py-0.5">gh auth login</code>{' '}
                in your terminal to authenticate.
              </p>
            )}
          </div>
        ) : null}
      </div>

      {/* Re-detect */}
      <div>
        <Button variant="outline" size="sm" onClick={detect} disabled={detecting}>
          <RefreshCw className={`mr-2 h-3.5 w-3.5 ${detecting ? 'animate-spin' : ''}`} />
          Re-detect
        </Button>
      </div>
    </div>
  )
}

function StatusRow({ label, value }: { label: string; value: React.ReactNode }): React.JSX.Element {
  return (
    <div className="flex items-center justify-between text-sm">
      <span className="text-muted-foreground">{label}</span>
      <span>{value}</span>
    </div>
  )
}

function LoadingRow({ label }: { label: string }): React.JSX.Element {
  return (
    <div className="flex items-center gap-3 text-sm text-muted-foreground">
      <Loader2 className="h-4 w-4 animate-spin" />
      {label}
    </div>
  )
}

function InstalledBadge({ installed }: { installed: boolean }): React.JSX.Element {
  return installed ? (
    <span className="flex items-center gap-1.5 text-emerald-500">
      <CheckCircle2 className="h-4 w-4" />
      Installed
    </span>
  ) : (
    <span className="flex items-center gap-1.5 text-destructive">
      <XCircle className="h-4 w-4" />
      Not installed
    </span>
  )
}
