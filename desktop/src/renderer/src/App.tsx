import { useCallback, useEffect, useState } from 'react'
import { ErrorBoundary } from './components/ErrorBoundary'
import { Sidebar, type HomeTab } from './components/Sidebar'
import { ClaudeCodeIntegration } from './pages/integrations/ClaudeCodeIntegration'
import { GitHubIntegration } from './pages/integrations/GitHubIntegration'
import { SkillsPage } from './pages/integrations/SkillsPage'
import { McpPage } from './pages/integrations/McpPage'
import { CliPage } from './pages/integrations/CliPage'
import { ProjectsPage } from './pages/projects/ProjectsPage'
import { ProjectChat } from './pages/projects/ProjectChat'
import { useProjectSessions, type OpenProject } from './hooks/useProjectSessions'

interface StoredProject {
  readonly id: string
  readonly name: string
  readonly path: string
  readonly githubUrl: string | null
  readonly createdAt: string
}

export function App(): React.JSX.Element {
  const [allProjects, setAllProjects] = useState<readonly StoredProject[]>([])
  const [homeTab, setHomeTab] = useState<HomeTab>('home')
  const sessions = useProjectSessions()

  const refreshProjects = useCallback(async () => {
    try {
      const result = await window.api.projects.list()
      setAllProjects(result.projects)
    } catch {
      setAllProjects([])
    }
  }, [])

  useEffect(() => {
    refreshProjects()
  }, [refreshProjects])

  const handleOpenProject = useCallback((project: OpenProject) => {
    sessions.openProject(project)
    refreshProjects()
  }, [sessions.openProject, refreshProjects])

  const handleGoHome = useCallback(() => {
    sessions.setActiveProject(null)
    setHomeTab('home')
  }, [sessions.setActiveProject])

  const handleSelectHomeTab = useCallback((tab: HomeTab) => {
    sessions.setActiveProject(null)
    setHomeTab(tab)
  }, [sessions.setActiveProject])

  const handleSelectProject = useCallback((id: string) => {
    sessions.setActivePanel(id, 'orchestrator')
    const alreadyOpen = sessions.openProjects.some((p) => p.id === id)
    if (!alreadyOpen) {
      const project = allProjects.find((p) => p.id === id)
      if (project) sessions.openProject(project)
      return
    }
    sessions.setActiveProject(id)
  }, [sessions.openProject, sessions.setActiveProject, sessions.setActivePanel, sessions.openProjects, allProjects])

  const handleSelectPanel = useCallback((panel: string) => {
    if (sessions.activeProjectId) {
      sessions.setActivePanel(sessions.activeProjectId, panel)
    }
  }, [sessions.activeProjectId, sessions.setActivePanel])

  const showingProject = sessions.activeProjectId !== null
  const currentPanel = sessions.activeProjectId
    ? sessions.getActivePanel(sessions.activeProjectId)
    : 'orchestrator'

  return (
    <div className="flex h-screen">
      <Sidebar
        onGoHome={handleGoHome}
        homeTab={homeTab}
        onSelectHomeTab={handleSelectHomeTab}
        allProjects={allProjects}
        openProjects={sessions.openProjects}
        activeProjectId={sessions.activeProjectId}
        lastActiveProjectId={sessions.lastActiveProjectId}
        statuses={sessions.statuses}
        onSelectProject={handleSelectProject}
        onSelectPanel={handleSelectPanel}
        activePanel={currentPanel}
      />

      <main className="flex flex-1 flex-col overflow-hidden">
        <ErrorBoundary>
        {!showingProject && (
          <div
            className="h-[52px] shrink-0"
            style={{ WebkitAppRegion: 'drag' } as React.CSSProperties}
          />
        )}

        {!showingProject && (
          <div className="flex-1 overflow-y-auto">
            {homeTab === 'home' && <ProjectsPage onOpenProject={handleOpenProject} />}
            {homeTab === 'claude-code' && <ClaudeCodeIntegration />}
            {homeTab === 'github' && <GitHubIntegration />}
            {homeTab === 'skills' && <SkillsPage />}
            {homeTab === 'mcp' && <McpPage />}
            {homeTab === 'cli' && <CliPage />}
          </div>
        )}

        {sessions.openProjects.map((project) => {
          const panel = sessions.getActivePanel(project.id)
          return (
            <ProjectChat
              key={project.id}
              project={project}
              hidden={project.id !== sessions.activeProjectId}
              activePanel={panel}
              onStatusChange={(status) => sessions.updateStatus(project.id, status)}
            />
          )
        })}
        </ErrorBoundary>
      </main>
    </div>
  )
}
