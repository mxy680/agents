import { useCallback, useEffect, useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  Terminal,
  RefreshCw,
  FolderOpen,
  CheckCircle2,
  XCircle,
  Loader2
} from 'lucide-react'

interface DetectionResult {
  readonly installed: boolean
  readonly path: string | null
  readonly version: string | null
  readonly error: string | null
}

export function ClaudeCodeIntegration(): React.JSX.Element {
  const [result, setResult] = useState<DetectionResult | null>(null)
  const [detecting, setDetecting] = useState(false)
  const [customPath, setCustomPath] = useState('')
  const [validating, setValidating] = useState(false)

  const detect = useCallback(async () => {
    setDetecting(true)
    try {
      const detection = await window.api.claudeCode.detect()
      setResult(detection)
    } catch (err) {
      setResult({
        installed: false,
        path: null,
        version: null,
        error: err instanceof Error ? err.message : 'Detection failed'
      })
    } finally {
      setDetecting(false)
    }
  }, [])

  useEffect(() => {
    detect()
  }, [detect])

  const handleValidate = async (): Promise<void> => {
    if (!customPath.trim()) return
    setValidating(true)
    try {
      const validation = await window.api.claudeCode.validate(customPath.trim())
      setResult(validation)
    } catch (err) {
      setResult({
        installed: false,
        path: customPath,
        version: null,
        error: err instanceof Error ? err.message : 'Validation failed'
      })
    } finally {
      setValidating(false)
    }
  }

  const isLoading = detecting || validating

  return (
    <div className="mx-auto max-w-2xl space-y-8 p-8">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-4">
          <div className="flex h-12 w-12 items-center justify-center rounded-lg bg-secondary">
            <Terminal className="h-6 w-6" />
          </div>
          <div>
            <h1 className="text-xl font-semibold">Claude Code</h1>
            <p className="text-sm text-muted-foreground">
              Anthropic's CLI for agentic development
            </p>
          </div>
        </div>
        {result && (
          <Badge variant={result.installed ? 'success' : 'destructive'}>
            {result.installed ? 'Connected' : 'Not Found'}
          </Badge>
        )}
      </div>

      {/* Status Card */}
      <div className="rounded-lg border bg-card p-6">
        <h2 className="mb-4 text-sm font-medium">Detection Status</h2>

        {isLoading ? (
          <div className="flex items-center gap-3 text-sm text-muted-foreground">
            <Loader2 className="h-4 w-4 animate-spin" />
            {detecting ? 'Detecting Claude Code...' : 'Validating path...'}
          </div>
        ) : result ? (
          <div className="space-y-3">
            <StatusRow
              label="Status"
              value={
                result.installed ? (
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
            />
            {result.path && <StatusRow label="Path" value={<code className="text-xs">{result.path}</code>} />}
            {result.version && <StatusRow label="Version" value={result.version} />}
            {result.error && (
              <StatusRow
                label="Error"
                value={<span className="text-destructive">{result.error}</span>}
              />
            )}
          </div>
        ) : null}

        <div className="mt-4">
          <Button variant="outline" size="sm" onClick={detect} disabled={isLoading}>
            <RefreshCw className={`mr-2 h-3.5 w-3.5 ${detecting ? 'animate-spin' : ''}`} />
            Re-detect
          </Button>
        </div>
      </div>

      {/* Custom Path */}
      <div className="rounded-lg border bg-card p-6">
        <h2 className="mb-1 text-sm font-medium">Custom Path</h2>
        <p className="mb-4 text-xs text-muted-foreground">
          Override auto-detection by specifying the path to the Claude Code binary.
        </p>

        <div className="flex gap-2">
          <div className="relative flex-1">
            <FolderOpen className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <input
              type="text"
              value={customPath}
              onChange={(e) => setCustomPath(e.target.value)}
              placeholder="/usr/local/bin/claude"
              className="h-9 w-full rounded-md border border-input bg-background pl-10 pr-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={handleValidate}
            disabled={!customPath.trim() || isLoading}
          >
            {validating ? <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" /> : null}
            Validate
          </Button>
        </div>
      </div>

      {/* Help */}
      {result && !result.installed && (
        <div className="rounded-lg border border-dashed bg-card/50 p-6 text-center">
          <p className="text-sm text-muted-foreground">
            Claude Code is not installed.{' '}
            <a
              href="https://docs.anthropic.com/en/docs/claude-code/overview"
              target="_blank"
              rel="noopener noreferrer"
              className="text-primary underline underline-offset-4 hover:text-primary/80"
            >
              Installation guide
            </a>
          </p>
        </div>
      )}
    </div>
  )
}

function StatusRow({
  label,
  value
}: {
  label: string
  value: React.ReactNode
}): React.JSX.Element {
  return (
    <div className="flex items-center justify-between text-sm">
      <span className="text-muted-foreground">{label}</span>
      <span>{value}</span>
    </div>
  )
}
