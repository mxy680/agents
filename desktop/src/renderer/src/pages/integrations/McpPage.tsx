import { useCallback, useEffect, useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import {
  Loader2,
  Server,
  Globe,
  FolderGit2,
  Search,
  ChevronDown,
  Terminal,
  Link,
  Trash2,
  KeyRound,
  CheckCircle2,
  AlertTriangle,
  XCircle,
  Circle,
  RefreshCw
} from 'lucide-react'

type McpStatus = 'connected' | 'needs-auth' | 'error' | 'unknown'

interface McpServer {
  readonly name: string
  readonly type: 'stdio' | 'http' | 'url' | 'unknown'
  readonly command?: string
  readonly url?: string
  readonly scope: 'global' | 'project'
  readonly source: string
  readonly sourcePath: string
  readonly status: McpStatus
}

interface Project {
  readonly id: string
  readonly name: string
  readonly path: string
}

export function McpPage(): React.JSX.Element {
  const [servers, setServers] = useState<readonly McpServer[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [projects, setProjects] = useState<readonly Project[]>([])
  const [selectedProjectId, setSelectedProjectId] = useState<string>('all')
  const [dropdownOpen, setDropdownOpen] = useState(false)

  useEffect(() => {
    window.api.projects.list().then((result) => {
      setProjects(result.projects)
    })
  }, [])

  const loadServers = useCallback(() => {
    setLoading(true)
    const selectedProject = projects.find((p) => p.id === selectedProjectId)
    window.api.mcp
      .list(selectedProject ? { projectPath: selectedProject.path } : undefined)
      .then((result) => {
        setServers(result.servers)
        setLoading(false)
      })
  }, [selectedProjectId, projects])

  useEffect(() => {
    loadServers()
  }, [loadServers])

  const handleRemove = async (server: McpServer): Promise<void> => {
    await window.api.mcp.remove({ serverName: server.name, sourcePath: server.sourcePath })
    loadServers()
  }

  const handleAuth = async (server: McpServer): Promise<void> => {
    await window.api.mcp.auth({ serverName: server.name })
    loadServers()
  }

  const globalServers = servers.filter((s) => s.scope === 'global')
  const projectServers = servers.filter((s) => s.scope === 'project')

  const filterServers = (list: readonly McpServer[]): readonly McpServer[] => {
    if (!search.trim()) return list
    const q = search.toLowerCase()
    return list.filter(
      (s) =>
        s.name.toLowerCase().includes(q) ||
        (s.command ?? '').toLowerCase().includes(q) ||
        (s.url ?? '').toLowerCase().includes(q)
    )
  }

  const filteredGlobal = filterServers(globalServers)
  const filteredProject = filterServers(projectServers)

  return (
    <div className="mx-auto max-w-2xl space-y-8 p-8">
      {/* Header */}
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-3">
          <div className="flex size-12 items-center justify-center rounded-lg bg-secondary">
            <Server className="size-6" />
          </div>
          <div>
            <h1 className="text-xl font-semibold">MCP Servers</h1>
            <p className="text-sm text-muted-foreground">
              Model Context Protocol servers for Claude Code
            </p>
          </div>
        </div>
        <Button variant="outline" size="sm" onClick={loadServers} disabled={loading}>
          <RefreshCw className={cn('mr-2 size-3.5', loading && 'animate-spin')} />
          Refresh
        </Button>
      </div>

      {/* Filters */}
      <div className="flex gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search servers..."
            className="h-9 w-full rounded-md border border-input bg-background pl-10 pr-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </div>

        <div className="relative">
          <button
            onClick={() => setDropdownOpen((v) => !v)}
            className="flex h-9 items-center gap-2 rounded-md border border-input bg-background px-3 text-sm hover:bg-accent"
          >
            <FolderGit2 className="size-3.5 text-muted-foreground" />
            <span className="max-w-[120px] truncate">
              {selectedProjectId === 'all'
                ? 'All projects'
                : projects.find((p) => p.id === selectedProjectId)?.name ?? 'Select'}
            </span>
            <ChevronDown className="size-3.5 text-muted-foreground" />
          </button>

          {dropdownOpen && (
            <div className="absolute right-0 top-10 z-50 w-48 rounded-md border bg-card py-1 shadow-lg">
              <button
                onClick={() => { setSelectedProjectId('all'); setDropdownOpen(false) }}
                className={cn(
                  'flex w-full items-center px-3 py-1.5 text-sm transition-colors hover:bg-accent',
                  selectedProjectId === 'all' && 'text-primary font-medium'
                )}
              >
                All projects
              </button>
              {projects.map((project) => (
                <button
                  key={project.id}
                  onClick={() => { setSelectedProjectId(project.id); setDropdownOpen(false) }}
                  className={cn(
                    'flex w-full items-center px-3 py-1.5 text-sm transition-colors hover:bg-accent',
                    selectedProjectId === project.id && 'text-primary font-medium'
                  )}
                >
                  {project.name}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-12 text-sm text-muted-foreground">
          <Loader2 className="mr-2 size-4 animate-spin" />
          Loading servers (checking status)...
        </div>
      ) : (
        <>
          {filteredProject.length > 0 && (
            <div>
              <div className="mb-3 flex items-center gap-2 text-sm font-medium">
                <FolderGit2 className="size-4 text-muted-foreground" />
                Project Servers
                <Badge variant="secondary" className="text-[10px]">{filteredProject.length}</Badge>
              </div>
              <div className="space-y-2">
                {filteredProject.map((server) => (
                  <ServerCard
                    key={`${server.scope}:${server.name}`}
                    server={server}
                    onRemove={() => handleRemove(server)}
                    onAuth={() => handleAuth(server)}
                  />
                ))}
              </div>
            </div>
          )}

          {filteredGlobal.length > 0 && (
            <div>
              <div className="mb-3 flex items-center gap-2 text-sm font-medium">
                <Globe className="size-4 text-muted-foreground" />
                Global Servers
                <Badge variant="secondary" className="text-[10px]">{filteredGlobal.length}</Badge>
              </div>
              <div className="space-y-2">
                {filteredGlobal.map((server) => (
                  <ServerCard
                    key={`${server.scope}:${server.source}:${server.name}`}
                    server={server}
                    onRemove={() => handleRemove(server)}
                    onAuth={() => handleAuth(server)}
                  />
                ))}
              </div>
            </div>
          )}

          {filteredGlobal.length === 0 && filteredProject.length === 0 && (
            <div className="py-8 text-center text-sm text-muted-foreground">
              {search ? 'No servers match your search.' : 'No MCP servers configured.'}
            </div>
          )}
        </>
      )}
    </div>
  )
}

function StatusIndicator({ status }: { status: McpStatus }): React.JSX.Element {
  switch (status) {
    case 'connected':
      return (
        <span className="flex items-center gap-1 text-emerald-500">
          <CheckCircle2 className="size-3.5" />
          <span className="text-[11px]">Connected</span>
        </span>
      )
    case 'needs-auth':
      return (
        <span className="flex items-center gap-1 text-amber-500">
          <AlertTriangle className="size-3.5" />
          <span className="text-[11px]">Needs auth</span>
        </span>
      )
    case 'error':
      return (
        <span className="flex items-center gap-1 text-destructive">
          <XCircle className="size-3.5" />
          <span className="text-[11px]">Error</span>
        </span>
      )
    default:
      return (
        <span className="flex items-center gap-1 text-muted-foreground">
          <Circle className="size-3.5" />
          <span className="text-[11px]">Unknown</span>
        </span>
      )
  }
}

function ServerCard({
  server,
  onRemove,
  onAuth
}: {
  server: McpServer
  onRemove: () => void
  onAuth: () => void
}): React.JSX.Element {
  const typeIcon = server.type === 'stdio'
    ? <Terminal className="size-4 text-primary" />
    : <Link className="size-4 text-primary" />

  return (
    <div className="group rounded-lg border bg-card p-4">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          {typeIcon}
          <span className="text-sm font-medium">{server.name}</span>
          <Badge variant="secondary" className="text-[10px]">{server.type}</Badge>
        </div>
        <div className="flex items-center gap-2">
          <StatusIndicator status={server.status} />
          <div className="flex items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100">
            {server.status === 'needs-auth' && (
              <button
                onClick={onAuth}
                className="rounded-md p-1 text-muted-foreground transition-colors hover:text-primary"
                title="Authenticate"
              >
                <KeyRound className="size-3.5" />
              </button>
            )}
            <button
              onClick={onRemove}
              className="rounded-md p-1 text-muted-foreground transition-colors hover:text-destructive"
              title="Remove server"
            >
              <Trash2 className="size-3.5" />
            </button>
          </div>
        </div>
      </div>
      <div className="mt-1.5 flex items-center justify-between">
        <p className="truncate text-xs text-muted-foreground">
          {server.command ? <code>{server.command}</code> : server.url}
        </p>
        <span className="shrink-0 text-[10px] text-muted-foreground/50">{server.source}</span>
      </div>
    </div>
  )
}
