import { useState, useCallback } from 'react'

export interface OpenProject {
  readonly id: string
  readonly name: string
  readonly path: string
  readonly githubUrl: string | null
  readonly createdAt: string
}

export type OrchestratorStatus = 'idle' | 'working'

export type MinionStatusType = 'spawning' | 'working' | 'done' | 'error'

export interface MinionSummary {
  readonly id: string
  readonly name: string
  readonly branch: string
  readonly status: MinionStatusType
}

export interface ProjectStatus {
  readonly orchestratorStatus: OrchestratorStatus
  readonly minionCount: number
  readonly activeMinionCount: number
  readonly minions: readonly MinionSummary[]
  readonly activePanel: string
}

const DEFAULT_STATUS: ProjectStatus = {
  orchestratorStatus: 'idle',
  minionCount: 0,
  activeMinionCount: 0,
  minions: [],
  activePanel: 'orchestrator'
}

// Per-project panel selection (stored separately so it doesn't trigger status updates)
const activePanels = new Map<string, string>()

export interface ProjectSessionsState {
  readonly openProjects: readonly OpenProject[]
  readonly activeProjectId: string | null
  readonly lastActiveProjectId: string | null
  readonly statuses: Readonly<Record<string, ProjectStatus>>
  openProject: (project: OpenProject) => void
  closeProject: (id: string) => void
  setActiveProject: (id: string | null) => void
  updateStatus: (id: string, status: ProjectStatus) => void
  getStatus: (id: string) => ProjectStatus
  setActivePanel: (projectId: string, panel: string) => void
  getActivePanel: (projectId: string) => string
}

export function useProjectSessions(): ProjectSessionsState {
  const [openProjects, setOpenProjects] = useState<readonly OpenProject[]>([])
  const [activeProjectId, setActiveProjectId] = useState<string | null>(null)
  const [lastActiveProjectId, setLastActiveProjectId] = useState<string | null>(null)
  const [statuses, setStatuses] = useState<Record<string, ProjectStatus>>({})

  const openProject = useCallback((project: OpenProject) => {
    setOpenProjects((prev) => {
      if (prev.some((p) => p.id === project.id)) return prev
      return [...prev, project]
    })
    setActiveProjectId(project.id)
    setLastActiveProjectId(project.id)
  }, [])

  const closeProject = useCallback((id: string) => {
    setOpenProjects((prev) => prev.filter((p) => p.id !== id))
    setStatuses((prev) => {
      const next = { ...prev }
      delete next[id]
      return next
    })
    setActiveProjectId((prev) => {
      if (prev !== id) return prev
      // Switch to another open project or null
      return null
    })
  }, [])

  const updateStatus = useCallback((id: string, status: ProjectStatus) => {
    setStatuses((prev) => ({ ...prev, [id]: status }))
  }, [])

  const getStatus = useCallback((id: string): ProjectStatus => {
    return statuses[id] ?? DEFAULT_STATUS
  }, [statuses])

  const setActiveProjectWrapped = useCallback((id: string | null) => {
    setActiveProjectId(id)
    if (id) setLastActiveProjectId(id)
  }, [])

  // Panel state uses a counter to trigger re-renders
  const [panelVersion, setPanelVersion] = useState(0)

  const setActivePanel = useCallback((projectId: string, panel: string) => {
    activePanels.set(projectId, panel)
    setPanelVersion((v) => v + 1)
  }, [])

  const getActivePanel = useCallback((projectId: string): string => {
    void panelVersion // read to subscribe to changes
    return activePanels.get(projectId) ?? 'orchestrator'
  }, [panelVersion])

  return {
    openProjects,
    activeProjectId,
    lastActiveProjectId,
    statuses,
    openProject,
    closeProject,
    setActiveProject: setActiveProjectWrapped,
    updateStatus,
    getStatus,
    setActivePanel,
    getActivePanel
  }
}
