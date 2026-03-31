import { useEffect, useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import {
  Loader2,
  Terminal,
  Search,
  Play,
  CheckCircle2,
  XCircle,
  ChevronDown,
  ChevronRight
} from 'lucide-react'

interface CliIntegration {
  readonly id: string
  readonly name: string
  readonly description: string
  readonly command: string
  readonly testCommand: string
  readonly requiresCredentials: boolean
}

type TestStatus = 'idle' | 'running' | 'pass' | 'fail'

interface TestResult {
  status: TestStatus
  output: string
  error: string | null
}

export function CliPage(): React.JSX.Element {
  const [integrations, setIntegrations] = useState<readonly CliIntegration[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [testResults, setTestResults] = useState<Record<string, TestResult>>({})

  useEffect(() => {
    window.api.cli.list().then((result) => {
      setIntegrations(result.integrations)
      setLoading(false)
    })
  }, [])

  const handleTest = async (id: string): Promise<void> => {
    setTestResults((prev) => ({
      ...prev,
      [id]: { status: 'running', output: '', error: null }
    }))

    const result = await window.api.cli.test({ integrationId: id })

    setTestResults((prev) => ({
      ...prev,
      [id]: {
        status: result.success ? 'pass' : 'fail',
        output: result.output,
        error: result.error
      }
    }))
  }

  const handleTestAll = async (): Promise<void> => {
    const credentialIntegrations = integrations.filter((i) => i.requiresCredentials)
    for (const integration of credentialIntegrations) {
      await handleTest(integration.id)
    }
  }

  const filtered = integrations.filter((i) =>
    !search.trim() ||
    i.name.toLowerCase().includes(search.toLowerCase()) ||
    i.description.toLowerCase().includes(search.toLowerCase()) ||
    i.command.toLowerCase().includes(search.toLowerCase())
  )

  const credentialIntegrations = filtered.filter((i) => i.requiresCredentials)
  const publicIntegrations = filtered.filter((i) => !i.requiresCredentials)

  const passCount = Object.values(testResults).filter((r) => r.status === 'pass').length
  const failCount = Object.values(testResults).filter((r) => r.status === 'fail').length
  const runningCount = Object.values(testResults).filter((r) => r.status === 'running').length

  return (
    <div className="mx-auto max-w-2xl space-y-8 p-8">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className="flex size-12 items-center justify-center rounded-lg bg-secondary">
            <Terminal className="size-6" />
          </div>
          <div>
            <h1 className="text-xl font-semibold">CLI Integrations</h1>
            <p className="text-sm text-muted-foreground">
              External services available via the <code>integrations</code> CLI
            </p>
          </div>
        </div>
        <Button variant="outline" size="sm" onClick={handleTestAll} disabled={runningCount > 0}>
          {runningCount > 0 ? (
            <Loader2 className="mr-2 size-3.5 animate-spin" />
          ) : (
            <Play className="mr-2 size-3.5" />
          )}
          Test All
        </Button>
      </div>

      {/* Summary */}
      {(passCount > 0 || failCount > 0) && (
        <div className="flex gap-3">
          {passCount > 0 && (
            <Badge variant="success" className="gap-1">
              <CheckCircle2 className="size-3" />
              {passCount} passed
            </Badge>
          )}
          {failCount > 0 && (
            <Badge variant="destructive" className="gap-1">
              <XCircle className="size-3" />
              {failCount} failed
            </Badge>
          )}
          {runningCount > 0 && (
            <Badge variant="default" className="gap-1">
              <Loader2 className="size-3 animate-spin" />
              {runningCount} running
            </Badge>
          )}
        </div>
      )}

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
        <input
          type="text"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search integrations..."
          className="h-9 w-full rounded-md border border-input bg-background pl-10 pr-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
        />
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-12 text-sm text-muted-foreground">
          <Loader2 className="mr-2 size-4 animate-spin" />
          Loading integrations...
        </div>
      ) : (
        <>
          {/* Credential-based */}
          {credentialIntegrations.length > 0 && (
            <div>
              <div className="mb-3 text-sm font-medium">
                Connected Services
                <Badge variant="secondary" className="ml-2 text-[10px]">{credentialIntegrations.length}</Badge>
              </div>
              <div className="space-y-2">
                {credentialIntegrations.map((integration) => (
                  <IntegrationCard
                    key={integration.id}
                    integration={integration}
                    result={testResults[integration.id]}
                    onTest={() => handleTest(integration.id)}
                  />
                ))}
              </div>
            </div>
          )}

          {/* Public APIs */}
          {publicIntegrations.length > 0 && (
            <div>
              <div className="mb-3 text-sm font-medium">
                Public APIs
                <Badge variant="secondary" className="ml-2 text-[10px]">{publicIntegrations.length}</Badge>
              </div>
              <div className="space-y-2">
                {publicIntegrations.map((integration) => (
                  <IntegrationCard
                    key={integration.id}
                    integration={integration}
                    result={testResults[integration.id]}
                    onTest={() => handleTest(integration.id)}
                  />
                ))}
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}

function IntegrationCard({
  integration,
  result,
  onTest
}: {
  integration: CliIntegration
  result?: TestResult
  onTest: () => void
}): React.JSX.Element {
  const [expanded, setExpanded] = useState(false)
  const status = result?.status ?? 'idle'

  return (
    <div className="group rounded-lg border bg-card p-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <StatusIcon status={status} />
          <div>
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium">{integration.name}</span>
              <code className="text-[10px] text-muted-foreground">{integration.command}</code>
            </div>
            <p className="text-xs text-muted-foreground">{integration.description}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {result && (result.output || result.error) && (
            <button
              onClick={() => setExpanded((v) => !v)}
              className="rounded-md p-1 text-muted-foreground transition-colors hover:text-foreground"
            >
              {expanded ? <ChevronDown className="size-4" /> : <ChevronRight className="size-4" />}
            </button>
          )}
          <Button
            variant="outline"
            size="sm"
            onClick={onTest}
            disabled={status === 'running'}
            className="h-7 text-xs"
          >
            {status === 'running' ? (
              <Loader2 className="mr-1.5 size-3 animate-spin" />
            ) : (
              <Play className="mr-1.5 size-3" />
            )}
            Test
          </Button>
        </div>
      </div>

      {/* Test command */}
      <div className="mt-2">
        <code className="text-[11px] text-muted-foreground/50">
          integrations {integration.testCommand}
        </code>
      </div>

      {/* Expanded output */}
      {expanded && result && (
        <div className="mt-3 space-y-2">
          {result.output && (
            <pre className="max-h-[200px] overflow-auto rounded bg-black/20 p-2 text-[11px] text-muted-foreground">
              {result.output}
            </pre>
          )}
          {result.error && (
            <pre className="max-h-[200px] overflow-auto rounded bg-destructive/10 p-2 text-[11px] text-destructive">
              {result.error}
            </pre>
          )}
        </div>
      )}
    </div>
  )
}

function StatusIcon({ status }: { status: TestStatus }): React.JSX.Element {
  switch (status) {
    case 'running':
      return <Loader2 className="size-5 animate-spin text-primary" />
    case 'pass':
      return <CheckCircle2 className="size-5 text-emerald-500" />
    case 'fail':
      return <XCircle className="size-5 text-destructive" />
    default:
      return <Terminal className="size-5 text-muted-foreground/30" />
  }
}
