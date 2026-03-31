import { spawn, ChildProcess, execFile } from 'child_process'
import { getEngagentEnv } from './engagent-config'
import { BrowserWindow, app } from 'electron'
import { randomUUID } from 'crypto'
import { writeFileSync } from 'fs'
import { promisify } from 'util'
import { join } from 'path'

const MINION_SYSTEM_PROMPT = `You are a minion agent — a specialized worker executing a specific task in your own git worktree.

# Workflow

1. Read your spec file (referenced in your task) thoroughly before writing any code
2. Implement all requirements from the spec
3. Verify your work using the commands listed in the spec
4. Write and run end-to-end tests

# Testing (MANDATORY)

After implementing any feature, you MUST write and run e2e tests. This is not optional.

## Process
1. Implement the feature first
2. Write e2e tests that cover:
   - Happy path for each user-facing flow
   - Error states and edge cases
   - Any integration points with other systems
3. Run the tests and fix any failures
4. Only report "done" when all tests pass

## Test Guidelines
- Use the project's existing test framework (check package.json for vitest, jest, playwright, etc.)
- If no test framework exists, install and configure one appropriate for the project
- Place tests next to the code they test, or in a __tests__ directory following the project's convention
- Test real behavior, not implementation details
- Mock external services but test real database/API interactions where possible
- Each test should be independent and not rely on other tests' state

## E2E Tests Specifically
- Test from the user's perspective (UI interactions, API calls)
- Cover the critical user journey described in the spec's acceptance criteria
- Use Playwright for browser-based e2e tests if the project has a frontend
- Use supertest or similar for API-only e2e tests
- Take screenshots on failure if using Playwright

# Git Practices
- Make atomic commits as you work
- Use conventional commit messages (feat:, fix:, test:, etc.)
- Commit tests separately from implementation when practical

# External Integrations
You have access to the \`integrations\` CLI for external services (Gmail, Calendar, GitHub, etc.). Use \`--json\` flag for machine-readable output. Credentials are resolved automatically.

# Completion
When done, ensure:
- All spec requirements are implemented
- All tests pass
- Code builds without errors
- Changes are committed to your branch`

let minionPromptFile: string | null = null

function getMinionPromptFile(): string {
  if (!minionPromptFile) {
    const filePath = join(app.getPath('userData'), 'minion-prompt.txt')
    writeFileSync(filePath, MINION_SYSTEM_PROMPT, 'utf-8')
    minionPromptFile = filePath
  }
  return minionPromptFile
}

const execFileAsync = promisify(execFile)
const shellPath = process.env.SHELL || '/bin/zsh'

async function runShell(command: string, cwd?: string): Promise<string> {
  const { stdout } = await execFileAsync(shellPath, ['-l', '-c', command], {
    timeout: 30000,
    cwd
  })
  return stdout.trim()
}

export type MinionStatus = 'spawning' | 'working' | 'done' | 'error'

export interface MinionInfo {
  readonly id: string
  readonly name: string
  readonly task: string
  readonly branch: string
  readonly worktreePath: string
  readonly projectPath: string
  readonly status: MinionStatus
  readonly createdAt: string
}

interface MinionInternal {
  id: string
  name: string
  task: string
  branch: string
  worktreePath: string
  projectPath: string
  status: MinionStatus
  createdAt: string
  process: ChildProcess | null
  claudeSessionId: string | null
}

const minions = new Map<string, MinionInternal>()

function sendMinionLog(
  window: BrowserWindow,
  minion: MinionInternal,
  role: 'assistant' | 'system',
  content: string
): void {
  window.webContents.send('minion:log', {
    minionId: minion.id,
    projectPath: minion.projectPath,
    log: {
      id: randomUUID(),
      role,
      content,
      timestamp: new Date().toISOString()
    }
  })
}

function getWorktreeDir(projectPath: string): string {
  return join(projectPath, '.ade-worktrees')
}

function sanitizeBranchName(name: string): string {
  return name.toLowerCase().replace(/[^a-z0-9-]/g, '-').replace(/-+/g, '-').replace(/^-|-$/g, '')
}

export async function spawnMinion(
  event: Electron.IpcMainInvokeEvent,
  input: { projectPath: string; name: string; task: string }
): Promise<{ minion: MinionInfo | null; error: string | null }> {
  const window = BrowserWindow.fromWebContents(event.sender)
  if (!window) {
    return { minion: null, error: 'Window not found' }
  }

  const id = randomUUID()
  const safeName = sanitizeBranchName(input.name)
  const branch = `minion/${safeName}`
  const worktreeBase = getWorktreeDir(input.projectPath)
  const worktreePath = join(worktreeBase, safeName)

  const minion: MinionInternal = {
    id,
    name: input.name,
    task: input.task,
    branch,
    worktreePath,
    projectPath: input.projectPath,
    status: 'spawning',
    createdAt: new Date().toISOString(),
    process: null,
    claudeSessionId: null
  }

  minions.set(id, minion)
  broadcastMinionUpdate(window, minion)

  try {
    // Create worktree directory
    await runShell(`mkdir -p "${worktreeBase}"`, input.projectPath)

    // Clean up stale worktree/branch from previous runs
    try {
      await runShell(`git worktree remove "${worktreePath}" --force`, input.projectPath)
    } catch { /* doesn't exist, fine */ }
    try {
      await runShell(`git branch -D "${branch}"`, input.projectPath)
    } catch { /* doesn't exist, fine */ }

    // Create the worktree with a new branch
    await runShell(
      `git worktree add "${worktreePath}" -b "${branch}" HEAD`,
      input.projectPath
    )

    minion.status = 'working'
    broadcastMinionUpdate(window, minion)

    // Start Claude Code in the worktree
    startMinionProcess(window, minion)

    return { minion: toMinionInfo(minion), error: null }
  } catch (err) {
    minion.status = 'error'
    broadcastMinionUpdate(window, minion)

    // Send error as a log entry
    sendMinionLog(window, minion, 'system',
      `Failed to create worktree: ${err instanceof Error ? err.message : 'Unknown error'}`
    )

    return {
      minion: null,
      error: err instanceof Error ? err.message : 'Failed to spawn minion'
    }
  }
}

function startMinionProcess(window: BrowserWindow, minion: MinionInternal, message?: string): void {
  const prompt = message ?? minion.task
  const promptFile = getMinionPromptFile()
  const args = [
    'claude',
    '-p',
    JSON.stringify(prompt),
    '--output-format', 'stream-json',
    '--verbose',
    '--dangerously-skip-permissions',
    '--append-system-prompt-file', `"${promptFile}"`
  ]

  if (minion.claudeSessionId) {
    args.push('--resume', minion.claudeSessionId)
  }

  const claudeCmd = args.join(' ')

  const child = spawn(shellPath, ['-l', '-c', claudeCmd + ' < /dev/null'], {
    cwd: minion.worktreePath,
    env: { ...process.env, ...getEngagentEnv() },
    stdio: ['ignore', 'pipe', 'pipe']
  })

  minion.process = child

  let buffer = ''

  child.stdout.on('data', (data: Buffer) => {
    buffer += data.toString()
    const lines = buffer.split('\n')
    buffer = lines.pop() || ''

    for (const line of lines) {
      const trimmed = line.trim()
      if (!trimmed) continue

      try {
        const parsed = JSON.parse(trimmed)

        // Capture session ID
        if (parsed.type === 'system' && parsed.subtype === 'init' && parsed.session_id) {
          minion.claudeSessionId = parsed.session_id
        }

        // Extract assistant messages (tool calls, text)
        if (parsed.type === 'assistant' && parsed.message) {
          const content = parsed.message.content
          if (Array.isArray(content)) {
            for (const block of content) {
              if (block.type === 'text' && block.text) {
                sendMinionLog(window, minion, 'assistant', block.text)
              } else if (block.type === 'tool_use') {
                const toolName = block.name || 'unknown'
                const inp = block.input || {}
                let desc = `**Tool: ${toolName}**`
                if (toolName === 'Bash' && inp.command) {
                  desc += `\n\`\`\`\n${inp.command}\n\`\`\``
                } else if ((toolName === 'Write' || toolName === 'Edit' || toolName === 'Read') && inp.file_path) {
                  desc += `\nFile: ${inp.file_path}`
                } else if ((toolName === 'Grep' || toolName === 'Glob') && inp.pattern) {
                  desc += `\nPattern: \`${inp.pattern}\``
                }
                sendMinionLog(window, minion, 'assistant', desc)
              }
            }
          }
        }

        // Tool results
        if (parsed.type === 'tool_result') {
          const content = parsed.content
          if (typeof content === 'string' && content.trim()) {
            const truncated = content.length > 500 ? content.slice(0, 500) + '...' : content
            sendMinionLog(window, minion, 'system', `\`\`\`\n${truncated}\n\`\`\``)
          }
        }

        // Forward raw events
        window.webContents.send('minion:stream', {
          minionId: minion.id,
          event: parsed
        })
      } catch {
        // Not valid JSON
      }
    }
  })

  child.stderr.on('data', (data: Buffer) => {
    const text = data.toString().trim()
    if (text) {
      sendMinionLog(window, minion, 'system', text)
    }
  })

  child.on('close', (code) => {
    minion.process = null
    minion.status = code === 0 ? 'done' : 'error'
    broadcastMinionUpdate(window, minion)
  })

  child.on('error', (err) => {
    minion.process = null
    minion.status = 'error'
    broadcastMinionUpdate(window, minion)

    sendMinionLog(window, minion, 'system', `Process error: ${err.message}`)
  })

  // Send initial log
  sendMinionLog(window, minion, 'system',
    `Assigned task: ${minion.task}\nWorking in branch: ${minion.branch}\nWorktree: ${minion.worktreePath}`
  )
}

export async function messageMinion(
  event: Electron.IpcMainInvokeEvent,
  input: { minionId: string; message: string }
): Promise<{ error: string | null }> {
  const minion = minions.get(input.minionId)
  if (!minion) return { error: 'Minion not found' }

  if (minion.process) {
    return { error: 'Minion is already working. Wait for it to finish or stop it first.' }
  }

  if (!minion.claudeSessionId) {
    return { error: 'Minion has no session to resume.' }
  }

  const window = BrowserWindow.fromWebContents(event.sender)
  if (!window) return { error: 'Window not found' }

  // Log the follow-up instruction
  sendMinionLog(window, minion, 'system', `Follow-up from orchestrator: ${input.message}`)

  minion.status = 'working'
  broadcastMinionUpdate(window, minion)

  // Start a new process resuming the existing session
  startMinionProcess(window, minion, input.message)

  return { error: null }
}

export function listMinions(): { minions: MinionInfo[] } {
  return { minions: Array.from(minions.values()).map(toMinionInfo) }
}

export function killMinion(
  _event: unknown,
  input: { minionId: string }
): { error: string | null } {
  const minion = minions.get(input.minionId)
  if (!minion) return { error: 'Minion not found' }

  if (minion.process) {
    minion.process.kill('SIGTERM')
    minion.process = null
  }
  minion.status = 'done'

  return { error: null }
}

export async function removeMinion(
  _event: unknown,
  input: { minionId: string; cleanupWorktree: boolean }
): Promise<{ error: string | null }> {
  const minion = minions.get(input.minionId)
  if (!minion) return { error: 'Minion not found' }

  // Kill process if running
  if (minion.process) {
    minion.process.kill('SIGTERM')
    minion.process = null
  }

  // Remove worktree and branch if requested
  if (input.cleanupWorktree) {
    try {
      await runShell(
        `git worktree remove "${minion.worktreePath}" --force`,
        minion.projectPath
      )
    } catch {
      // Worktree may already be gone
    }
    try {
      await runShell(
        `git branch -D "${minion.branch}"`,
        minion.projectPath
      )
    } catch {
      // Branch may already be gone
    }
  }

  minions.delete(input.minionId)
  return { error: null }
}

function toMinionInfo(m: MinionInternal): MinionInfo {
  return {
    id: m.id,
    name: m.name,
    task: m.task,
    branch: m.branch,
    worktreePath: m.worktreePath,
    projectPath: m.projectPath,
    status: m.status,
    createdAt: m.createdAt
  }
}

function broadcastMinionUpdate(window: BrowserWindow, minion: MinionInternal): void {
  window.webContents.send('minion:update', toMinionInfo(minion))
}
