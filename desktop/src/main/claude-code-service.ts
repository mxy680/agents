import { execFile } from 'child_process'
import { access, constants } from 'fs/promises'
import { promisify } from 'util'

const execFileAsync = promisify(execFile)

export interface ClaudeCodeDetectionResult {
  readonly installed: boolean
  readonly path: string | null
  readonly version: string | null
  readonly error: string | null
}

const KNOWN_PATHS = [
  '/usr/local/bin/claude',
  '/opt/homebrew/bin/claude',
  `${process.env.HOME}/.npm-global/bin/claude`,
  `${process.env.HOME}/.local/bin/claude`,
  `${process.env.HOME}/.claude/local/claude`
]

async function isExecutable(filePath: string): Promise<boolean> {
  try {
    await access(filePath, constants.X_OK)
    return true
  } catch {
    return false
  }
}

async function getVersion(claudePath: string): Promise<string | null> {
  try {
    const { stdout } = await execFileAsync(claudePath, ['--version'], { timeout: 5000 })
    return stdout.trim()
  } catch {
    return null
  }
}

async function findViaShell(): Promise<string | null> {
  const shell = process.env.SHELL || '/bin/zsh'
  try {
    const { stdout } = await execFileAsync(shell, ['-l', '-c', 'which claude'], { timeout: 5000 })
    const path = stdout.trim()
    return path || null
  } catch {
    return null
  }
}

async function findViaKnownPaths(): Promise<string | null> {
  for (const knownPath of KNOWN_PATHS) {
    if (await isExecutable(knownPath)) {
      return knownPath
    }
  }
  return null
}

export async function detectClaudeCode(): Promise<ClaudeCodeDetectionResult> {
  try {
    const path = (await findViaShell()) ?? (await findViaKnownPaths())

    if (!path) {
      return { installed: false, path: null, version: null, error: null }
    }

    const version = await getVersion(path)

    return { installed: true, path, version, error: null }
  } catch (err) {
    return {
      installed: false,
      path: null,
      version: null,
      error: err instanceof Error ? err.message : 'Unknown error'
    }
  }
}

export async function validateClaudeCodePath(
  customPath: string
): Promise<ClaudeCodeDetectionResult> {
  try {
    if (!(await isExecutable(customPath))) {
      return {
        installed: false,
        path: customPath,
        version: null,
        error: 'File is not executable'
      }
    }

    const version = await getVersion(customPath)

    if (!version) {
      return {
        installed: false,
        path: customPath,
        version: null,
        error: 'Not a valid Claude Code binary'
      }
    }

    return { installed: true, path: customPath, version, error: null }
  } catch (err) {
    return {
      installed: false,
      path: customPath,
      version: null,
      error: err instanceof Error ? err.message : 'Unknown error'
    }
  }
}
