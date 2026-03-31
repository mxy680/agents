import { app, shell, BrowserWindow, ipcMain } from 'electron'
app.name = 'ade'

process.on('uncaughtException', (err) => {
  console.error('[CRASH] Uncaught exception:', err.message, err.stack)
})
process.on('unhandledRejection', (reason) => {
  console.error('[CRASH] Unhandled rejection:', reason)
})
import { join } from 'path'
import { electronApp, optimizer, is } from '@electron-toolkit/utils'
import { detectClaudeCode, validateClaudeCodePath } from './claude-code-service'
import {
  detectGitHub,
  listGitHubRepos,
  createGitHubRepo,
  pushProjectToGitHub
} from './github-service'
import {
  spawnMinion,
  messageMinion,
  listMinions,
  killMinion,
  removeMinion
} from './minion-service'
import {
  createChatSession,
  sendChatMessage,
  abortChatMessage,
  destroyChatSession
} from './chat-service'
import { saveSession, loadSession, deleteSession } from './session-store'
import { listCliIntegrations, testCliIntegration } from './cli-integrations-service'
import { getEngagentConfigHandler, saveEngagentConfigHandler } from './engagent-config'
import { listSkills, removeSkill } from './skills-service'
import { listMcpServers, removeMcpServer, authMcpServer } from './mcp-service'
import {
  listProjects,
  getDefaultDir,
  createProject,
  connectProject,
  updateProjectGithubUrl,
  removeProject
} from './projects-service'

function createWindow(): void {
  const mainWindow = new BrowserWindow({
    width: 960,
    height: 670,
    minWidth: 720,
    minHeight: 480,
    titleBarStyle: 'hiddenInset',
    trafficLightPosition: { x: 16, y: 16 },
    show: false,
    webPreferences: {
      preload: join(__dirname, '../preload/index.js'),
      sandbox: false
    }
  })

  mainWindow.on('ready-to-show', () => {
    mainWindow.show()
  })

  // Log renderer crashes and errors
  mainWindow.webContents.on('render-process-gone', (_event, details) => {
    console.error('[CRASH] Renderer process gone:', details.reason, details.exitCode)
  })

  mainWindow.webContents.on('did-fail-load', (_event, errorCode, errorDescription) => {
    console.error('[CRASH] Failed to load:', errorCode, errorDescription)
  })

  mainWindow.webContents.on('console-message', (_event, level, message, line, sourceId) => {
    if (level >= 2) { // warnings and errors
      console.error(`[RENDERER ${level === 2 ? 'WARN' : 'ERROR'}] ${message} (${sourceId}:${line})`)
    }
  })

  mainWindow.webContents.setWindowOpenHandler((details) => {
    shell.openExternal(details.url)
    return { action: 'deny' }
  })

  if (is.dev && process.env['ELECTRON_RENDERER_URL']) {
    mainWindow.loadURL(process.env['ELECTRON_RENDERER_URL'])
  } else {
    mainWindow.loadFile(join(__dirname, '../renderer/index.html'))
  }
}

app.whenReady().then(() => {
  electronApp.setAppUserModelId('com.emdash.ade')

  app.on('browser-window-created', (_, window) => {
    optimizer.watchWindowShortcuts(window)
  })

  ipcMain.handle('claude-code:detect', detectClaudeCode)
  ipcMain.handle('claude-code:validate', (_event, path: string) => validateClaudeCodePath(path))
  ipcMain.handle('github:detect', detectGitHub)
  ipcMain.handle('github:repos', listGitHubRepos)
  ipcMain.handle('github:create-repo', createGitHubRepo)
  ipcMain.handle('github:push', pushProjectToGitHub)
  ipcMain.handle('projects:list', listProjects)
  ipcMain.handle('projects:default-dir', getDefaultDir)
  ipcMain.handle('projects:create', createProject)
  ipcMain.handle('projects:connect', connectProject)
  ipcMain.handle('projects:update-github-url', updateProjectGithubUrl)
  ipcMain.handle('projects:remove', removeProject)
  ipcMain.handle('minion:spawn', spawnMinion)
  ipcMain.handle('minion:message', messageMinion)
  ipcMain.handle('minion:list', listMinions)
  ipcMain.handle('minion:kill', killMinion)
  ipcMain.handle('minion:remove', removeMinion)
  ipcMain.handle('cli:list', listCliIntegrations)
  ipcMain.handle('cli:test', testCliIntegration)
  ipcMain.handle('engagent:get-config', getEngagentConfigHandler)
  ipcMain.handle('engagent:save-config', saveEngagentConfigHandler)
  ipcMain.handle('skills:list', listSkills)
  ipcMain.handle('mcp:list', listMcpServers)
  ipcMain.handle('mcp:remove', removeMcpServer)
  ipcMain.handle('mcp:auth', authMcpServer)
  ipcMain.handle('skills:remove', removeSkill)
  ipcMain.handle('session:save', saveSession)
  ipcMain.handle('session:load', loadSession)
  ipcMain.handle('session:delete', deleteSession)
  ipcMain.handle('chat:create-session', createChatSession)
  ipcMain.handle('chat:send', sendChatMessage)
  ipcMain.handle('chat:abort', abortChatMessage)
  ipcMain.handle('chat:destroy-session', destroyChatSession)

  createWindow()

  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) createWindow()
  })
})

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit()
  }
})
