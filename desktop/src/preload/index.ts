import { contextBridge, ipcRenderer } from 'electron'
import { electronAPI } from '@electron-toolkit/preload'

export interface ClaudeCodeDetectionResult {
  readonly installed: boolean
  readonly path: string | null
  readonly version: string | null
  readonly error: string | null
}

const api = {
  claudeCode: {
    detect: (): Promise<ClaudeCodeDetectionResult> => ipcRenderer.invoke('claude-code:detect'),
    validate: (path: string): Promise<ClaudeCodeDetectionResult> =>
      ipcRenderer.invoke('claude-code:validate', path)
  },
  github: {
    detect: (): Promise<unknown> => ipcRenderer.invoke('github:detect'),
    listRepos: (limit?: number): Promise<unknown> => ipcRenderer.invoke('github:repos', limit),
    createRepo: (input: {
      name: string
      description: string
      isPrivate: boolean
    }): Promise<unknown> => ipcRenderer.invoke('github:create-repo', input),
    push: (input: {
      projectPath: string
      repoName: string
      isPrivate: boolean
    }): Promise<unknown> => ipcRenderer.invoke('github:push', input)
  },
  projects: {
    list: (): Promise<unknown> => ipcRenderer.invoke('projects:list'),
    defaultDir: (): Promise<string> => ipcRenderer.invoke('projects:default-dir'),
    create: (input: { name: string; parentDir: string }): Promise<unknown> =>
      ipcRenderer.invoke('projects:create', input),
    connect: (input: { githubUrl: string; parentDir: string }): Promise<unknown> =>
      ipcRenderer.invoke('projects:connect', input),
    updateGithubUrl: (input: { projectId: string; githubUrl: string }): Promise<unknown> =>
      ipcRenderer.invoke('projects:update-github-url', input),
    remove: (input: {
      projectId: string
      deleteFiles: boolean
      deleteRepo: boolean
    }): Promise<unknown> => ipcRenderer.invoke('projects:remove', input)
  },
  minion: {
    spawn: (input: { projectPath: string; name: string; task: string }): Promise<unknown> =>
      ipcRenderer.invoke('minion:spawn', input),
    list: (): Promise<unknown> => ipcRenderer.invoke('minion:list'),
    message: (input: { minionId: string; message: string }): Promise<unknown> =>
      ipcRenderer.invoke('minion:message', input),
    kill: (input: { minionId: string }): Promise<unknown> =>
      ipcRenderer.invoke('minion:kill', input),
    remove: (input: { minionId: string; cleanupWorktree: boolean }): Promise<unknown> =>
      ipcRenderer.invoke('minion:remove', input),
    onUpdate: (callback: (data: unknown) => void): (() => void) => {
      const handler = (_event: unknown, data: unknown): void => callback(data)
      ipcRenderer.on('minion:update', handler)
      return () => ipcRenderer.removeListener('minion:update', handler)
    },
    onLog: (callback: (data: { minionId: string; projectPath: string; log: unknown }) => void): (() => void) => {
      const handler = (_event: unknown, data: { minionId: string; projectPath: string; log: unknown }): void => callback(data)
      ipcRenderer.on('minion:log', handler)
      return () => ipcRenderer.removeListener('minion:log', handler)
    }
  },
  cli: {
    list: (): Promise<unknown> => ipcRenderer.invoke('cli:list'),
    test: (input: { integrationId: string }): Promise<unknown> => ipcRenderer.invoke('cli:test', input)
  },
  mcp: {
    list: (input?: { projectPath?: string }): Promise<unknown> =>
      ipcRenderer.invoke('mcp:list', input),
    remove: (input: { serverName: string; sourcePath: string }): Promise<unknown> =>
      ipcRenderer.invoke('mcp:remove', input),
    auth: (input: { serverName: string }): Promise<unknown> =>
      ipcRenderer.invoke('mcp:auth', input)
  },
  skills: {
    list: (input?: { projectPath?: string }): Promise<unknown> =>
      ipcRenderer.invoke('skills:list', input),
    remove: (input: { skillPath: string }): Promise<unknown> =>
      ipcRenderer.invoke('skills:remove', input)
  },
  session: {
    save: (input: { projectId: string; data: unknown }): Promise<unknown> =>
      ipcRenderer.invoke('session:save', input),
    load: (input: { projectId: string }): Promise<unknown> =>
      ipcRenderer.invoke('session:load', input),
    delete: (input: { projectId: string }): Promise<unknown> =>
      ipcRenderer.invoke('session:delete', input)
  },
  chat: {
    createSession: (input: { projectPath: string; claudeSessionId?: string }): Promise<{ sessionId: string }> =>
      ipcRenderer.invoke('chat:create-session', input),
    send: (input: { sessionId: string; message: string }): Promise<{ error: string | null }> =>
      ipcRenderer.invoke('chat:send', input),
    abort: (input: { sessionId: string }): Promise<{ error: string | null }> =>
      ipcRenderer.invoke('chat:abort', input),
    destroySession: (input: { sessionId: string }): Promise<void> =>
      ipcRenderer.invoke('chat:destroy-session', input),
    onStream: (callback: (data: { sessionId: string; event: unknown }) => void): (() => void) => {
      const handler = (_event: unknown, data: { sessionId: string; event: unknown }): void =>
        callback(data)
      ipcRenderer.on('chat:stream', handler)
      return () => ipcRenderer.removeListener('chat:stream', handler)
    },
    onDone: (callback: (data: { sessionId: string }) => void): (() => void) => {
      const handler = (_event: unknown, data: { sessionId: string }): void => callback(data)
      ipcRenderer.on('chat:done', handler)
      return () => ipcRenderer.removeListener('chat:done', handler)
    }
  }
}

if (process.contextIsolated) {
  try {
    contextBridge.exposeInMainWorld('electron', electronAPI)
    contextBridge.exposeInMainWorld('api', api)
  } catch (error) {
    console.error(error)
  }
} else {
  // @ts-expect-error global augmentation
  window.electron = electronAPI
  // @ts-expect-error global augmentation
  window.api = api
}
