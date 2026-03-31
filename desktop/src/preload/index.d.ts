import { ElectronAPI } from '@electron-toolkit/preload'

interface ClaudeCodeDetectionResult {
  readonly installed: boolean
  readonly path: string | null
  readonly version: string | null
  readonly error: string | null
}

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

interface Project {
  readonly id: string
  readonly name: string
  readonly path: string
  readonly githubUrl: string | null
  readonly createdAt: string
}

interface ProjectsResult {
  readonly projects: readonly Project[]
}

interface ProjectResult {
  readonly project: Project | null
  readonly error: string | null
}

interface GitHubRepo {
  readonly nameWithOwner: string
  readonly description: string
  readonly isPrivate: boolean
  readonly url: string
  readonly updatedAt: string
}

interface GitHubReposResult {
  readonly repos: readonly GitHubRepo[]
  readonly error: string | null
}

interface CreateGitHubRepoResult {
  readonly url: string | null
  readonly error: string | null
}

interface PushToGitHubResult {
  readonly url: string | null
  readonly error: string | null
}

type MinionStatus = 'spawning' | 'working' | 'done' | 'error'

interface MinionInfo {
  readonly id: string
  readonly name: string
  readonly task: string
  readonly branch: string
  readonly worktreePath: string
  readonly projectPath: string
  readonly status: MinionStatus
  readonly createdAt: string
}

interface MinionLog {
  readonly id: string
  readonly role: 'assistant' | 'system'
  readonly content: string
  readonly timestamp: string
}

interface Api {
  claudeCode: {
    detect: () => Promise<ClaudeCodeDetectionResult>
    validate: (path: string) => Promise<ClaudeCodeDetectionResult>
  }
  github: {
    detect: () => Promise<GitHubDetectionResult>
    listRepos: (limit?: number) => Promise<GitHubReposResult>
    createRepo: (input: {
      name: string
      description: string
      isPrivate: boolean
    }) => Promise<CreateGitHubRepoResult>
    push: (input: {
      projectPath: string
      repoName: string
      isPrivate: boolean
    }) => Promise<PushToGitHubResult>
  }
  projects: {
    list: () => Promise<ProjectsResult>
    defaultDir: () => Promise<string>
    create: (input: { name: string; parentDir: string }) => Promise<ProjectResult>
    connect: (input: { githubUrl: string; parentDir: string }) => Promise<ProjectResult>
    updateGithubUrl: (input: { projectId: string; githubUrl: string }) => Promise<{ error: string | null }>
    remove: (input: {
      projectId: string
      deleteFiles: boolean
      deleteRepo: boolean
    }) => Promise<{ error: string | null }>
  }
  minion: {
    spawn: (input: { projectPath: string; name: string; task: string }) => Promise<{ minion: MinionInfo | null; error: string | null }>
    list: () => Promise<{ minions: MinionInfo[] }>
    message: (input: { minionId: string; message: string }) => Promise<{ error: string | null }>
    kill: (input: { minionId: string }) => Promise<{ error: string | null }>
    remove: (input: { minionId: string; cleanupWorktree: boolean }) => Promise<{ error: string | null }>
    onUpdate: (callback: (minion: MinionInfo) => void) => () => void
    onLog: (callback: (data: { minionId: string; projectPath: string; log: MinionLog }) => void) => () => void
  }
  cli: {
    list: () => Promise<{ integrations: readonly { id: string; name: string; description: string; command: string; testCommand: string; requiresCredentials: boolean }[] }>
    test: (input: { integrationId: string }) => Promise<{ success: boolean; output: string; error: string | null }>
  }
  mcp: {
    list: (input?: { projectPath?: string }) => Promise<{
      servers: readonly {
        name: string
        type: 'stdio' | 'http' | 'url' | 'unknown'
        command?: string
        url?: string
        scope: 'global' | 'project'
        source: string
        sourcePath: string
        status: 'connected' | 'needs-auth' | 'error' | 'unknown'
      }[]
      error: string | null
    }>
    remove: (input: { serverName: string; sourcePath: string }) => Promise<{ error: string | null }>
    auth: (input: { serverName: string }) => Promise<{ error: string | null }>
  }
  skills: {
    list: (input?: { projectPath?: string }) => Promise<{
      skills: readonly {
        name: string
        description: string
        origin: string
        scope: 'global' | 'project'
        path: string
      }[]
      error: string | null
    }>
    remove: (input: { skillPath: string }) => Promise<{ error: string | null }>
  }
  session: {
    save: (input: { projectId: string; data: unknown }) => Promise<{ error: string | null }>
    load: (input: { projectId: string }) => Promise<{ data: unknown; error: string | null }>
    delete: (input: { projectId: string }) => Promise<{ error: string | null }>
  }
  chat: {
    createSession: (input: { projectPath: string; claudeSessionId?: string }) => Promise<{ sessionId: string }>
    send: (input: { sessionId: string; message: string }) => Promise<{ error: string | null }>
    abort: (input: { sessionId: string }) => Promise<{ error: string | null }>
    destroySession: (input: { sessionId: string }) => Promise<void>
    onStream: (callback: (data: { sessionId: string; event: Record<string, unknown> }) => void) => () => void
    onDone: (callback: (data: { sessionId: string }) => void) => () => void
  }
}

declare global {
  interface Window {
    electron: ElectronAPI
    api: Api
  }
}
