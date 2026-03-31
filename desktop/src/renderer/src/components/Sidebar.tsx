import { FolderGit2, Loader2, Circle, Crown, Cpu, Home, Terminal, GitBranch, Zap, Server } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { OpenProject, ProjectStatus, MinionStatusType } from '@/hooks/useProjectSessions'

interface StoredProject {
  readonly id: string
  readonly name: string
  readonly path: string
  readonly githubUrl: string | null
  readonly createdAt: string
}

export type HomeTab = 'home' | 'claude-code' | 'github' | 'skills' | 'mcp' | 'cli'

interface SidebarProps {
  onGoHome: () => void
  homeTab: HomeTab
  onSelectHomeTab: (tab: HomeTab) => void
  allProjects: readonly StoredProject[]
  openProjects: readonly OpenProject[]
  activeProjectId: string | null
  lastActiveProjectId: string | null
  statuses: Readonly<Record<string, ProjectStatus>>
  onSelectProject: (id: string) => void
  onSelectPanel: (panel: string) => void
  activePanel: string
}

function SidebarButton({
  icon,
  label,
  active,
  onClick
}: {
  icon: React.ReactNode
  label: string
  active: boolean
  onClick: () => void
}): React.JSX.Element {
  return (
    <button
      onClick={onClick}
      className={cn(
        'flex w-full items-center gap-3 rounded-md px-2 py-1.5 text-sm transition-colors',
        active
          ? 'bg-sidebar-accent text-sidebar-accent-foreground'
          : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground'
      )}
    >
      {icon}
      {label}
    </button>
  )
}

export function Sidebar({
  onGoHome,
  homeTab,
  onSelectHomeTab,
  allProjects,
  openProjects,
  activeProjectId,
  lastActiveProjectId,
  statuses,
  onSelectProject,
  onSelectPanel,
  activePanel
}: SidebarProps): React.JSX.Element {
  const openIds = new Set(openProjects.map((p) => p.id))

  // Show orchestrator/minions for the last active project (even when on Home/integrations)
  const contextProjectId = lastActiveProjectId && openIds.has(lastActiveProjectId) ? lastActiveProjectId : null
  const contextProject = allProjects.find((p) => p.id === contextProjectId)
  const contextStatus = contextProjectId ? statuses[contextProjectId] : null
  const contextMinions = contextStatus?.minions ?? []
  const isOrchestratorWorking = contextStatus?.orchestratorStatus === 'working'

  return (
    <aside className="flex h-full w-[240px] flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground">
      <div className="h-[52px] shrink-0" style={{ WebkitAppRegion: 'drag' } as React.CSSProperties} />

      <nav className="flex-1 overflow-y-auto px-3 pb-4">
        <button
          onClick={onGoHome}
          className={cn(
            'mb-3 flex w-full items-center gap-3 rounded-md px-2 py-1.5 text-sm transition-colors',
            !activeProjectId
              ? 'bg-sidebar-accent text-sidebar-accent-foreground'
              : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground'
          )}
        >
          <Home className="size-4" />
          Home
        </button>

        {/* Projects */}
        <div className="mb-1 px-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          Projects
        </div>

        {allProjects.length > 0 ? (
          <div className="space-y-0.5">
            {allProjects.map((project) => {
              const isOpen = openIds.has(project.id)
              const isActive = project.id === activeProjectId
              const status = statuses[project.id]
              const isWorking = status?.orchestratorStatus === 'working'

              return (
                <button
                  key={project.id}
                  onClick={() => onSelectProject(project.id)}
                  className={cn(
                    'flex w-full items-center gap-2.5 rounded-md px-2 py-1.5 text-sm transition-colors',
                    isActive
                      ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                      : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50 hover:text-sidebar-foreground'
                  )}
                >
                  {isOpen ? (
                    isWorking ? (
                      <Loader2 className="size-3 animate-spin text-primary" />
                    ) : (
                      <Circle className={cn(
                        'size-2.5 fill-current',
                        (status?.activeMinionCount ?? 0) > 0
                          ? 'text-blue-400 animate-pulse'
                          : 'text-emerald-400'
                      )} />
                    )
                  ) : (
                    <Circle className="size-2.5 text-muted-foreground/30" />
                  )}
                  <span className="flex-1 truncate text-left">{project.name}</span>
                  {isOpen && (status?.minionCount ?? 0) > 0 && (
                    <span className="rounded-full bg-muted px-1.5 text-[10px] text-muted-foreground">
                      {status?.minionCount}
                    </span>
                  )}
                </button>
              )
            })}
          </div>
        ) : (
          <button
            onClick={onGoHome}
            className="w-full rounded-md px-2 py-1.5 text-left text-xs text-muted-foreground hover:text-sidebar-foreground"
          >
            No projects yet
          </button>
        )}

        {/* Orchestrator */}
        {contextProjectId && (
          <>
            <div className="mb-1 mt-6 px-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              <span>Orchestrator</span>
              {contextProject && !activeProjectId && (
                <span className="ml-1 font-normal normal-case tracking-normal text-muted-foreground/50">
                  — {contextProject.name}
                </span>
              )}
            </div>

            <button
              onClick={() => onSelectPanel('orchestrator')}
              className={cn(
                'flex w-full items-center gap-3 rounded-md px-2 py-2 text-sm transition-colors',
                activePanel === 'orchestrator'
                  ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                  : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50'
              )}
            >
              <div className="flex size-7 items-center justify-center rounded-md bg-primary/15">
                <Crown className="size-3.5 text-primary" />
              </div>
              <div className="flex-1 text-left">
                <div className="font-medium">Coordinator</div>
              </div>
              {isOrchestratorWorking && (
                <Loader2 className="size-3 animate-spin text-muted-foreground" />
              )}
            </button>
          </>
        )}

        {/* Minions */}
        {contextProjectId && (
          <>
            <div className="mb-1 mt-6 px-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
              <div className="flex items-center gap-2">
                <span>Minions</span>
                <span className="rounded-full bg-muted px-1.5 py-0.5 text-[10px]">
                  {contextMinions.length}
                </span>
              </div>
            </div>

            {contextMinions.length === 0 && (
              <p className="px-2 py-1 text-xs text-muted-foreground/50">No minions yet</p>
            )}
            {contextMinions.map((minion) => (
              <button
                key={minion.id}
                onClick={() => onSelectPanel(minion.id)}
                className={cn(
                  'flex w-full items-center gap-3 rounded-md px-2 py-2 text-sm transition-colors',
                  activePanel === minion.id
                    ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                    : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/50'
                )}
              >
                <div className="flex size-7 items-center justify-center rounded-md bg-secondary">
                  <Cpu className="size-3.5" />
                </div>
                <div className="flex-1 text-left min-w-0">
                  <div className="font-medium truncate">{minion.name}</div>
                  <div className="truncate text-[11px] text-muted-foreground">{minion.branch}</div>
                </div>
                <MinionStatusDot status={minion.status} />
              </button>
            ))}
          </>
        )}

        {/* Integrations — always visible */}
        <div className="mb-1 mt-6 px-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          Integrations
        </div>

        <SidebarButton
          icon={<Terminal className="size-4" />}
          label="Claude Code"
          active={!activeProjectId && homeTab === 'claude-code'}
          onClick={() => onSelectHomeTab('claude-code')}
        />
        <SidebarButton
          icon={<GitBranch className="size-4" />}
          label="GitHub"
          active={!activeProjectId && homeTab === 'github'}
          onClick={() => onSelectHomeTab('github')}
        />
        <SidebarButton
          icon={<Zap className="size-4" />}
          label="Skills"
          active={!activeProjectId && homeTab === 'skills'}
          onClick={() => onSelectHomeTab('skills')}
        />
        <SidebarButton
          icon={<Server className="size-4" />}
          label="MCP Servers"
          active={!activeProjectId && homeTab === 'mcp'}
          onClick={() => onSelectHomeTab('mcp')}
        />
        <SidebarButton
          icon={<Terminal className="size-4" />}
          label="CLI"
          active={!activeProjectId && homeTab === 'cli'}
          onClick={() => onSelectHomeTab('cli')}
        />
      </nav>
    </aside>
  )
}

function MinionStatusDot({ status }: { status: MinionStatusType }): React.JSX.Element {
  const colors: Record<MinionStatusType, string> = {
    spawning: 'text-yellow-400 animate-pulse',
    working: 'text-blue-400 animate-pulse',
    done: 'text-emerald-400',
    error: 'text-destructive'
  }
  return <Circle className={cn('size-2.5 fill-current', colors[status])} />
}
