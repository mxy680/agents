import { execFile } from 'child_process'
import { readFile, writeFile } from 'fs/promises'
import { join } from 'path'
import { homedir } from 'os'
import { promisify } from 'util'

const execFileAsync = promisify(execFile)
const shellPath = process.env.SHELL || '/bin/zsh'

export type McpStatus = 'connected' | 'needs-auth' | 'error' | 'unknown'

export interface McpServer {
  readonly name: string
  readonly type: 'stdio' | 'http' | 'url' | 'unknown'
  readonly command?: string
  readonly url?: string
  readonly scope: 'global' | 'project'
  readonly source: string
  readonly sourcePath: string
  readonly status: McpStatus
}

interface McpResult {
  readonly servers: readonly McpServer[]
  readonly error: string | null
}

async function loadJsonFile(path: string): Promise<Record<string, unknown> | null> {
  try {
    const raw = await readFile(path, 'utf-8')
    return JSON.parse(raw)
  } catch {
    return null
  }
}

function extractServers(
  data: Record<string, unknown> | null,
  scope: 'global' | 'project',
  source: string,
  sourcePath: string
): McpServer[] {
  if (!data) return []

  const mcpServers = data.mcpServers as Record<string, Record<string, unknown>> | undefined
  if (!mcpServers || typeof mcpServers !== 'object') return []

  return Object.entries(mcpServers).map(([name, config]) => {
    const type = (config.type as string) ??
      (config.command ? 'stdio' : config.url ? 'url' : 'unknown')

    return {
      name,
      type: type as McpServer['type'],
      command: config.command as string | undefined,
      url: config.url as string | undefined,
      scope,
      source,
      sourcePath,
      status: 'unknown' as McpStatus
    }
  })
}

// Get live statuses from claude CLI init event
async function getServerStatuses(): Promise<Record<string, McpStatus>> {
  try {
    const { stdout } = await execFileAsync(
      shellPath,
      ['-l', '-c', 'claude -p "." --output-format stream-json --verbose --dangerously-skip-permissions --no-session-persistence < /dev/null 2>&1 | head -1'],
      { timeout: 30000 }
    )
    const parsed = JSON.parse(stdout.trim())
    const statuses: Record<string, McpStatus> = {}
    if (Array.isArray(parsed.mcp_servers)) {
      for (const s of parsed.mcp_servers) {
        statuses[s.name] = s.status as McpStatus
      }
    }
    return statuses
  } catch {
    return {}
  }
}

export async function listMcpServers(
  _event: unknown,
  input?: { projectPath?: string }
): Promise<McpResult> {
  try {
    const home = homedir()

    const globalClaudeJsonPath = join(home, '.claude.json')
    const globalClaudeJson = await loadJsonFile(globalClaudeJsonPath)
    const globalServers = extractServers(globalClaudeJson, 'global', '~/.claude.json', globalClaudeJsonPath)

    const globalSettingsPath = join(home, '.claude', 'settings.json')
    const globalSettings = await loadJsonFile(globalSettingsPath)
    const settingsServers = extractServers(globalSettings, 'global', '~/.claude/settings.json', globalSettingsPath)

    let projectServers: McpServer[] = []
    if (input?.projectPath) {
      const p1 = join(input.projectPath, '.claude', 'settings.local.json')
      const d1 = await loadJsonFile(p1)
      projectServers = extractServers(d1, 'project', '.claude/settings.local.json', p1)

      const p2 = join(input.projectPath, '.claude.json')
      const d2 = await loadJsonFile(p2)
      projectServers = [
        ...projectServers,
        ...extractServers(d2, 'project', '.claude.json', p2)
      ]
    }

    // Get live statuses
    const statuses = await getServerStatuses()
    const allServers = [...projectServers, ...globalServers, ...settingsServers].map((s) => ({
      ...s,
      status: statuses[s.name] ?? 'unknown'
    }))

    return { servers: allServers, error: null }
  } catch (err) {
    return {
      servers: [],
      error: err instanceof Error ? err.message : 'Failed to list MCP servers'
    }
  }
}

export async function removeMcpServer(
  _event: unknown,
  input: { serverName: string; sourcePath: string }
): Promise<{ error: string | null }> {
  try {
    const raw = await readFile(input.sourcePath, 'utf-8')
    const data = JSON.parse(raw)

    if (data.mcpServers && data.mcpServers[input.serverName]) {
      delete data.mcpServers[input.serverName]
      await writeFile(input.sourcePath, JSON.stringify(data, null, 2), 'utf-8')
    }

    return { error: null }
  } catch (err) {
    return { error: err instanceof Error ? err.message : 'Failed to remove server' }
  }
}

export async function authMcpServer(
  _event: unknown,
  input: { serverName: string }
): Promise<{ error: string | null }> {
  try {
    await execFileAsync(
      shellPath,
      ['-l', '-c', `claude mcp auth "${input.serverName}"`],
      { timeout: 60000 }
    )
    return { error: null }
  } catch (err) {
    return { error: err instanceof Error ? err.message : 'Failed to authenticate' }
  }
}
