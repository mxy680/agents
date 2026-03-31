import { execFile } from 'child_process'
import { promisify } from 'util'

const execFileAsync = promisify(execFile)

export interface GitDetectionResult {
  readonly installed: boolean
  readonly path: string | null
  readonly version: string | null
  readonly error: string | null
}

export interface GhAuthResult {
  readonly authenticated: boolean
  readonly username: string | null
  readonly scopes: readonly string[]
  readonly error: string | null
}

export interface GitHubDetectionResult {
  readonly git: GitDetectionResult
  readonly gh: GitDetectionResult
  readonly auth: GhAuthResult
}

const shell = process.env.SHELL || '/bin/zsh'

async function runShell(command: string): Promise<string | null> {
  try {
    const { stdout } = await execFileAsync(shell, ['-l', '-c', command], { timeout: 10000 })
    return stdout.trim() || null
  } catch {
    return null
  }
}

async function detectGit(): Promise<GitDetectionResult> {
  try {
    const path = await runShell('which git')
    if (!path) {
      return { installed: false, path: null, version: null, error: null }
    }

    const version = await runShell('git --version')

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

async function detectGh(): Promise<GitDetectionResult> {
  try {
    const path = await runShell('which gh')
    if (!path) {
      return { installed: false, path: null, version: null, error: null }
    }

    const version = await runShell('gh --version')
    const firstLine = version?.split('\n')[0] ?? null

    return { installed: true, path, version: firstLine, error: null }
  } catch (err) {
    return {
      installed: false,
      path: null,
      version: null,
      error: err instanceof Error ? err.message : 'Unknown error'
    }
  }
}

async function detectAuth(): Promise<GhAuthResult> {
  try {
    const status = await runShell('gh auth status 2>&1')
    if (!status) {
      return { authenticated: false, username: null, scopes: [], error: null }
    }

    const isAuthenticated = status.includes('Logged in to')

    let username: string | null = null
    const accountMatch = status.match(/Logged in to [^ ]+ account ([^ ]+)/)
      ?? status.match(/Logged in to [^ ]+ as ([^ ]+)/)
    if (accountMatch) {
      username = accountMatch[1].replace(/\(.*\)/, '').trim()
    }

    let scopes: string[] = []
    const scopesMatch = status.match(/Token scopes: (.+)/)
    if (scopesMatch) {
      scopes = scopesMatch[1].split(',').map((s) => s.trim().replace(/^'|'$/g, ''))
    }

    return { authenticated: isAuthenticated, username, scopes, error: null }
  } catch (err) {
    return {
      authenticated: false,
      username: null,
      scopes: [],
      error: err instanceof Error ? err.message : 'Unknown error'
    }
  }
}

export async function detectGitHub(): Promise<GitHubDetectionResult> {
  const [git, gh, auth] = await Promise.all([detectGit(), detectGh(), detectAuth()])
  return { git, gh, auth }
}

export interface GitHubRepo {
  readonly nameWithOwner: string
  readonly description: string
  readonly isPrivate: boolean
  readonly url: string
  readonly updatedAt: string
}

export interface ListReposResult {
  readonly repos: readonly GitHubRepo[]
  readonly error: string | null
}

export async function listGitHubRepos(
  _event: unknown,
  limit: number = 30
): Promise<ListReposResult> {
  try {
    const output = await runShell(
      `gh repo list --limit ${limit} --json nameWithOwner,description,isPrivate,url,updatedAt`
    )
    if (!output) {
      return { repos: [], error: 'Not authenticated. Run `gh auth login` first.' }
    }
    const repos: GitHubRepo[] = JSON.parse(output)
    return { repos, error: null }
  } catch (err) {
    return {
      repos: [],
      error: err instanceof Error ? err.message : 'Failed to list repos'
    }
  }
}

export interface CreateGitHubRepoInput {
  readonly name: string
  readonly description: string
  readonly isPrivate: boolean
}

export interface CreateGitHubRepoResult {
  readonly url: string | null
  readonly error: string | null
}

export async function createGitHubRepo(
  _event: unknown,
  input: CreateGitHubRepoInput
): Promise<CreateGitHubRepoResult> {
  try {
    const visibility = input.isPrivate ? '--private' : '--public'
    const descFlag = input.description ? ` --description "${input.description.replace(/"/g, '\\"')}"` : ''
    const output = await runShell(
      `gh repo create "${input.name}" ${visibility}${descFlag} --confirm 2>&1`
    )
    if (!output) {
      return { url: null, error: 'Failed to create repository' }
    }
    const urlMatch = output.match(/(https:\/\/github\.com\/[^\s]+)/)
    return { url: urlMatch ? urlMatch[1] : null, error: null }
  } catch (err) {
    return {
      url: null,
      error: err instanceof Error ? err.message : 'Failed to create repo'
    }
  }
}

export interface PushToGitHubResult {
  readonly url: string | null
  readonly error: string | null
}

export async function pushProjectToGitHub(
  _event: unknown,
  input: { projectPath: string; repoName: string; isPrivate: boolean }
): Promise<PushToGitHubResult> {
  try {
    const visibility = input.isPrivate ? '--private' : '--public'
    const dir = input.projectPath

    // Check if remote already exists in the project directory
    const existingRemote = await runShell(`git -C "${dir}" remote get-url origin 2>/dev/null`)

    if (existingRemote) {
      await runShellStrict(`git -C "${dir}" push -u origin HEAD`)
      return { url: existingRemote, error: null }
    }

    // Check if repo name already exists on GitHub
    const repoExists = await runShell(
      `gh repo view "${input.repoName}" --json url 2>/dev/null`
    )
    if (repoExists) {
      return {
        url: null,
        error: `Repository "${input.repoName}" already exists on GitHub. Use a different project name.`
      }
    }

    // Step 1: Create the repo on GitHub (without --source to avoid auto-push)
    const createOutput = await runShellStrict(
      `gh repo create "${input.repoName}" ${visibility} 2>&1`
    )
    const urlMatch = createOutput.match(/(https:\/\/github\.com\/[^\s]+)/)
    const repoUrl = urlMatch ? urlMatch[1] : null

    if (!repoUrl) {
      return { url: null, error: 'Repo created but could not parse URL from output' }
    }

    // Step 2: Add HTTPS remote and push (avoids SSH key issues)
    await runShellStrict(`git -C "${dir}" remote add origin "${repoUrl}.git"`)
    await runShellStrict(`git -C "${dir}" push -u origin HEAD`)

    return { url: repoUrl, error: null }
  } catch (err) {
    return {
      url: null,
      error: err instanceof Error ? err.message : 'Failed to push to GitHub'
    }
  }
}

async function runShellStrict(command: string): Promise<string> {
  const { stdout } = await execFileAsync(shell, ['-l', '-c', command], { timeout: 60000 })
  return stdout.trim()
}
