import { execFile } from 'child_process'
import { app } from 'electron'
import { access, mkdir, readFile, writeFile } from 'fs/promises'
import { join } from 'path'
import { promisify } from 'util'

const execFileAsync = promisify(execFile)
const shell = process.env.SHELL || '/bin/zsh'

export interface Project {
  readonly id: string
  readonly name: string
  readonly path: string
  readonly githubUrl: string | null
  readonly createdAt: string
}

export interface CreateProjectInput {
  readonly name: string
  readonly parentDir: string
}

export interface ConnectProjectInput {
  readonly githubUrl: string
  readonly parentDir: string
}

export interface ProjectsResult {
  readonly projects: readonly Project[]
}

export interface ProjectResult {
  readonly project: Project | null
  readonly error: string | null
}

function getConfigPath(): string {
  return join(app.getPath('userData'), 'projects.json')
}

function getDefaultProjectsDir(): string {
  return join(app.getPath('home'), 'ADE')
}

async function loadProjects(): Promise<Project[]> {
  try {
    const data = await readFile(getConfigPath(), 'utf-8')
    const parsed = JSON.parse(data)
    return Array.isArray(parsed.projects) ? parsed.projects : []
  } catch {
    return []
  }
}

async function saveProjects(projects: readonly Project[]): Promise<void> {
  const configPath = getConfigPath()
  const dir = join(configPath, '..')
  await mkdir(dir, { recursive: true })
  await writeFile(configPath, JSON.stringify({ projects }, null, 2), 'utf-8')
}

async function runShell(command: string, cwd?: string): Promise<string> {
  const { stdout } = await execFileAsync(shell, ['-l', '-c', command], {
    timeout: 60000,
    cwd
  })
  return stdout.trim()
}

function generateId(): string {
  return Date.now().toString(36) + Math.random().toString(36).slice(2, 8)
}

function extractRepoName(githubUrl: string): string {
  const cleaned = githubUrl.replace(/\.git$/, '').replace(/\/$/, '')
  const parts = cleaned.split('/')
  return parts[parts.length - 1] || 'project'
}

export async function listProjects(): Promise<ProjectsResult> {
  const projects = await loadProjects()
  return { projects }
}

export async function getDefaultDir(): Promise<string> {
  return getDefaultProjectsDir()
}

export async function createProject(
  _event: unknown,
  input: CreateProjectInput
): Promise<ProjectResult> {
  try {
    const projectDir = join(input.parentDir, input.name)

    // Check if directory already exists
    try {
      await access(projectDir)
      return { project: null, error: `Directory "${projectDir}" already exists.` }
    } catch {
      // Directory doesn't exist — good
    }

    // Check if project name is already tracked
    const existing = await loadProjects()
    if (existing.some((p) => p.name === input.name)) {
      return { project: null, error: `A project named "${input.name}" already exists.` }
    }

    await mkdir(projectDir, { recursive: true })
    await runShell('git init', projectDir)
    await runShell('git commit --allow-empty -m "Initial commit"', projectDir)

    const project: Project = {
      id: generateId(),
      name: input.name,
      path: projectDir,
      githubUrl: null,
      createdAt: new Date().toISOString()
    }

    await saveProjects([...existing, project])

    return { project, error: null }
  } catch (err) {
    return {
      project: null,
      error: err instanceof Error ? err.message : 'Failed to create project'
    }
  }
}

export async function connectProject(
  _event: unknown,
  input: ConnectProjectInput
): Promise<ProjectResult> {
  try {
    const repoName = extractRepoName(input.githubUrl)
    const projectDir = join(input.parentDir, repoName)

    await mkdir(input.parentDir, { recursive: true })
    await runShell(`git clone "${input.githubUrl}" "${projectDir}"`)

    const project: Project = {
      id: generateId(),
      name: repoName,
      path: projectDir,
      githubUrl: input.githubUrl,
      createdAt: new Date().toISOString()
    }

    const existing = await loadProjects()
    await saveProjects([...existing, project])

    return { project, error: null }
  } catch (err) {
    return {
      project: null,
      error: err instanceof Error ? err.message : 'Failed to clone project'
    }
  }
}

export async function updateProjectGithubUrl(
  _event: unknown,
  input: { projectId: string; githubUrl: string }
): Promise<{ error: string | null }> {
  try {
    const existing = await loadProjects()
    const updated = existing.map((p) =>
      p.id === input.projectId ? { ...p, githubUrl: input.githubUrl } : p
    )
    await saveProjects(updated)
    return { error: null }
  } catch (err) {
    return { error: err instanceof Error ? err.message : 'Failed to update project' }
  }
}

export async function removeProject(
  _event: unknown,
  input: { projectId: string; deleteFiles: boolean; deleteRepo: boolean }
): Promise<{ error: string | null }> {
  try {
    const existing = await loadProjects()
    const project = existing.find((p) => p.id === input.projectId)

    if (!project) {
      return { error: 'Project not found' }
    }

    // Delete GitHub repo if requested
    if (input.deleteRepo && project.githubUrl) {
      const repoMatch = project.githubUrl.match(/github\.com\/([^/]+\/[^/]+?)(?:\.git)?$/)
      if (repoMatch) {
        try {
          await runShell(`gh repo delete "${repoMatch[1]}" --yes 2>&1`)
        } catch (deleteErr) {
          const msg = deleteErr instanceof Error ? deleteErr.message : ''
          if (msg.includes('delete_repo')) {
            return {
              error: 'GitHub token lacks the "delete_repo" scope. Run: gh auth refresh -h github.com -s delete_repo'
            }
          }
          return { error: `Failed to delete GitHub repo: ${msg}` }
        }
      }
    }

    // Delete local files if requested
    if (input.deleteFiles) {
      const { rm } = await import('fs/promises')
      await rm(project.path, { recursive: true, force: true })
    }

    const filtered = existing.filter((p) => p.id !== input.projectId)
    await saveProjects(filtered)
    return { error: null }
  } catch (err) {
    return { error: err instanceof Error ? err.message : 'Failed to remove project' }
  }
}
