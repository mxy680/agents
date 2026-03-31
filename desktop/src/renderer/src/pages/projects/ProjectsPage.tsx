import { useCallback, useEffect, useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import {
  FolderOpen,
  Plus,
  Link,
  Trash2,
  Loader2,
  FolderGit2,
  ExternalLink,
  Clock,
  Upload,
  Lock,
  Globe,
  Search,
  Check,
  AlertTriangle
} from 'lucide-react'

interface Project {
  readonly id: string
  readonly name: string
  readonly path: string
  readonly githubUrl: string | null
  readonly createdAt: string
}

interface GitHubRepo {
  readonly nameWithOwner: string
  readonly description: string
  readonly isPrivate: boolean
  readonly url: string
  readonly updatedAt: string
}

type View = 'list' | 'create' | 'connect'

interface ProjectsPageProps {
  onOpenProject?: (project: Project) => void
}

export function ProjectsPage({ onOpenProject }: ProjectsPageProps): React.JSX.Element {
  const [projects, setProjects] = useState<readonly Project[]>([])
  const [loading, setLoading] = useState(true)
  const [view, setView] = useState<View>('list')
  const [deleteTarget, setDeleteTarget] = useState<Project | null>(null)

  const refresh = useCallback(async () => {
    setLoading(true)
    try {
      const result = await window.api.projects.list()
      setProjects(result.projects)
    } catch {
      setProjects([])
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    refresh()
  }, [refresh])

  const handleCreated = (): void => {
    setView('list')
    refresh()
  }

  const handleRemoveConfirmed = async (
    projectId: string,
    deleteFiles: boolean,
    deleteRepo: boolean
  ): Promise<string | null> => {
    const result = await window.api.projects.remove({ projectId, deleteFiles, deleteRepo })
    if (result.error) {
      return result.error
    }
    setDeleteTarget(null)
    refresh()
    return null
  }

  if (view === 'create') {
    return <CreateProjectView onBack={() => setView('list')} onCreated={handleCreated} />
  }

  if (view === 'connect') {
    return <ConnectProjectView onBack={() => setView('list')} onCreated={handleCreated} />
  }

  return (
    <div className="flex h-full flex-col items-center justify-center gap-8 p-8">
      <div className="text-center">
        <h1 className="text-2xl font-semibold">ADE</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Agentic Development Environment
        </p>
      </div>

      <div className="flex gap-4">
        <button
          onClick={() => setView('connect')}
          className="flex w-[200px] flex-col items-center gap-3 rounded-xl border bg-card p-6 text-center transition-colors hover:bg-accent"
        >
          <Link className="size-8 text-muted-foreground" />
          <div>
            <p className="text-sm font-medium">Connect Repo</p>
            <p className="mt-0.5 text-xs text-muted-foreground">Clone from GitHub</p>
          </div>
        </button>

        <button
          onClick={() => setView('create')}
          className="flex w-[200px] flex-col items-center gap-3 rounded-xl border bg-card p-6 text-center transition-colors hover:bg-accent"
        >
          <Plus className="size-8 text-muted-foreground" />
          <div>
            <p className="text-sm font-medium">New Project</p>
            <p className="mt-0.5 text-xs text-muted-foreground">Create a git repo</p>
          </div>
        </button>
      </div>
    </div>
  )
}

function ProjectCard({
  project,
  onOpen,
  onRemove,
  onPushed
}: {
  project: Project
  onOpen: () => void
  onRemove: () => void
  onPushed: () => void
}): React.JSX.Element {
  const [pushing, setPushing] = useState(false)
  const [pushError, setPushError] = useState<string | null>(null)

  const handlePush = async (): Promise<void> => {
    setPushing(true)
    setPushError(null)
    try {
      const result = await window.api.github.push({
        projectPath: project.path,
        repoName: project.name,
        isPrivate: true
      })
      if (result.error) {
        setPushError(result.error)
      } else {
        if (result.url) {
          await window.api.projects.updateGithubUrl({
            projectId: project.id,
            githubUrl: result.url
          })
        }
        onPushed()
      }
    } catch (err) {
      setPushError(err instanceof Error ? err.message : 'Push failed')
    } finally {
      setPushing(false)
    }
  }

  return (
    <div
      className="group cursor-pointer rounded-lg border bg-card p-4 transition-colors hover:bg-card/80"
      onClick={onOpen}
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-secondary">
            <FolderGit2 className="h-5 w-5" />
          </div>
          <div>
            <div className="flex items-center gap-2">
              <span className="text-sm font-medium">{project.name}</span>
              {project.githubUrl && (
                <Badge variant="secondary" className="text-[10px]">
                  GitHub
                </Badge>
              )}
            </div>
            <div className="mt-0.5 flex items-center gap-3 text-xs text-muted-foreground">
              <span className="flex items-center gap-1">
                <FolderOpen className="h-3 w-3" />
                {project.path}
              </span>
              <span className="flex items-center gap-1">
                <Clock className="h-3 w-3" />
                {new Date(project.createdAt).toLocaleDateString()}
              </span>
            </div>
            {project.githubUrl && (
              <div className="mt-0.5 flex items-center gap-1 text-xs text-muted-foreground">
                <ExternalLink className="h-3 w-3" />
                {project.githubUrl}
              </div>
            )}
          </div>
        </div>
        <div className="flex items-center gap-1 opacity-0 transition-opacity group-hover:opacity-100" onClick={(e) => e.stopPropagation()}>
          {!project.githubUrl && (
            <Button
              variant="ghost"
              size="sm"
              className="h-8 text-xs"
              onClick={handlePush}
              disabled={pushing}
            >
              {pushing ? (
                <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin" />
              ) : (
                <Upload className="mr-1.5 h-3.5 w-3.5" />
              )}
              Push to GitHub
            </Button>
          )}
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8"
            onClick={onRemove}
          >
            <Trash2 className="h-4 w-4 text-muted-foreground" />
          </Button>
        </div>
      </div>
      {pushError && (
        <p className="mt-2 text-xs text-destructive">{pushError}</p>
      )}
    </div>
  )
}

function CreateProjectView({
  onBack,
  onCreated
}: {
  onBack: () => void
  onCreated: () => void
}): React.JSX.Element {
  const [name, setName] = useState('')
  const [parentDir, setParentDir] = useState('')
  const [creating, setCreating] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [createOnGitHub, setCreateOnGitHub] = useState(false)
  const [isPrivate, setIsPrivate] = useState(true)

  useEffect(() => {
    window.api.projects.defaultDir().then(setParentDir)
  }, [])

  const handleCreate = async (): Promise<void> => {
    if (!name.trim()) return
    setCreating(true)
    setError(null)
    try {
      // Create local project
      const result = await window.api.projects.create({
        name: name.trim(),
        parentDir
      })
      if (result.error) {
        setError(result.error)
        return
      }

      // Optionally create on GitHub and push
      if (createOnGitHub && result.project) {
        const pushResult = await window.api.github.push({
          projectPath: result.project.path,
          repoName: name.trim(),
          isPrivate
        })
        if (pushResult.error) {
          setError(`Project created locally but GitHub push failed: ${pushResult.error}`)
          return
        }
        if (pushResult.url) {
          await window.api.projects.updateGithubUrl({
            projectId: result.project.id,
            githubUrl: pushResult.url
          })
        }
      }

      onCreated()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create project')
    } finally {
      setCreating(false)
    }
  }

  return (
    <div className="mx-auto max-w-2xl space-y-8 p-8">
      <div>
        <button
          onClick={onBack}
          className="mb-4 text-xs text-muted-foreground hover:text-foreground"
        >
          &larr; Back to projects
        </button>
        <h1 className="text-xl font-semibold">New Project</h1>
        <p className="text-sm text-muted-foreground">
          Create a new project with a local git repository.
        </p>
      </div>

      <div className="space-y-6 rounded-lg border bg-card p-6">
        <div>
          <label className="mb-1.5 block text-sm font-medium">Project Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="my-project"
            className="h-9 w-full rounded-md border border-input bg-background px-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
            onKeyDown={(e) => e.key === 'Enter' && handleCreate()}
          />
        </div>

        <div>
          <label className="mb-1.5 block text-sm font-medium">Location</label>
          <div className="relative">
            <FolderOpen className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <input
              type="text"
              value={parentDir}
              onChange={(e) => setParentDir(e.target.value)}
              className="h-9 w-full rounded-md border border-input bg-background pl-10 pr-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>
          {name.trim() && (
            <p className="mt-1.5 text-xs text-muted-foreground">
              Will create: <code>{parentDir}/{name.trim()}</code>
            </p>
          )}
        </div>

        {/* GitHub toggle */}
        <div className="rounded-md border p-4">
          <label className="flex cursor-pointer items-center justify-between">
            <div className="flex items-center gap-3">
              <Upload className="h-4 w-4 text-muted-foreground" />
              <div>
                <p className="text-sm font-medium">Create on GitHub</p>
                <p className="text-xs text-muted-foreground">
                  Also create a GitHub repository and push
                </p>
              </div>
            </div>
            <ToggleSwitch checked={createOnGitHub} onChange={setCreateOnGitHub} />
          </label>

          {createOnGitHub && (
            <div className="mt-4 flex gap-2">
              <button
                onClick={() => setIsPrivate(true)}
                className={`flex items-center gap-1.5 rounded-md border px-3 py-1.5 text-xs transition-colors ${
                  isPrivate
                    ? 'border-primary bg-primary/10 text-primary'
                    : 'border-input text-muted-foreground hover:text-foreground'
                }`}
              >
                <Lock className="h-3 w-3" />
                Private
              </button>
              <button
                onClick={() => setIsPrivate(false)}
                className={`flex items-center gap-1.5 rounded-md border px-3 py-1.5 text-xs transition-colors ${
                  !isPrivate
                    ? 'border-primary bg-primary/10 text-primary'
                    : 'border-input text-muted-foreground hover:text-foreground'
                }`}
              >
                <Globe className="h-3 w-3" />
                Public
              </button>
            </div>
          )}
        </div>

        {error && (
          <p className="text-sm text-destructive">{error}</p>
        )}

        <div className="flex justify-end gap-2">
          <Button variant="outline" size="sm" onClick={onBack}>
            Cancel
          </Button>
          <Button size="sm" onClick={handleCreate} disabled={!name.trim() || creating}>
            {creating && <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" />}
            {createOnGitHub ? 'Create & Push' : 'Create Project'}
          </Button>
        </div>
      </div>
    </div>
  )
}

function ConnectProjectView({
  onBack,
  onCreated
}: {
  onBack: () => void
  onCreated: () => void
}): React.JSX.Element {
  const [githubUrl, setGithubUrl] = useState('')
  const [parentDir, setParentDir] = useState('')
  const [connecting, setConnecting] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [mode, setMode] = useState<'url' | 'browse'>('browse')
  const [repos, setRepos] = useState<readonly GitHubRepo[]>([])
  const [loadingRepos, setLoadingRepos] = useState(false)
  const [reposError, setReposError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [selectedRepo, setSelectedRepo] = useState<GitHubRepo | null>(null)

  useEffect(() => {
    window.api.projects.defaultDir().then(setParentDir)
  }, [])

  useEffect(() => {
    if (mode === 'browse') {
      loadRepos()
    }
  }, [mode])

  const loadRepos = async (): Promise<void> => {
    setLoadingRepos(true)
    setReposError(null)
    try {
      const result = await window.api.github.listRepos(50)
      if (result.error) {
        setReposError(result.error)
      } else {
        setRepos(result.repos)
      }
    } catch {
      setReposError('Failed to load repositories')
    } finally {
      setLoadingRepos(false)
    }
  }

  const filteredRepos = repos.filter((r) =>
    r.nameWithOwner.toLowerCase().includes(search.toLowerCase()) ||
    (r.description && r.description.toLowerCase().includes(search.toLowerCase()))
  )

  const effectiveUrl = mode === 'browse' && selectedRepo ? selectedRepo.url : githubUrl.trim()
  const repoName = effectiveUrl
    .replace(/\.git$/, '')
    .replace(/\/$/, '')
    .split('/')
    .pop() || ''

  const handleConnect = async (): Promise<void> => {
    if (!effectiveUrl) return
    setConnecting(true)
    setError(null)
    try {
      const result = await window.api.projects.connect({
        githubUrl: effectiveUrl,
        parentDir
      })
      if (result.error) {
        setError(result.error)
      } else {
        onCreated()
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to connect project')
    } finally {
      setConnecting(false)
    }
  }

  return (
    <div className="mx-auto max-w-2xl space-y-8 p-8">
      <div>
        <button
          onClick={onBack}
          className="mb-4 text-xs text-muted-foreground hover:text-foreground"
        >
          &larr; Back to projects
        </button>
        <h1 className="text-xl font-semibold">Connect Repository</h1>
        <p className="text-sm text-muted-foreground">
          Clone a GitHub repository to your local machine.
        </p>
      </div>

      {/* Mode tabs */}
      <div className="flex gap-1 rounded-lg border bg-muted p-1">
        <button
          onClick={() => { setMode('browse'); setSelectedRepo(null) }}
          className={`flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
            mode === 'browse'
              ? 'bg-background text-foreground shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          Browse Your Repos
        </button>
        <button
          onClick={() => { setMode('url'); setSelectedRepo(null) }}
          className={`flex-1 rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
            mode === 'url'
              ? 'bg-background text-foreground shadow-sm'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          Paste URL
        </button>
      </div>

      <div className="space-y-6 rounded-lg border bg-card p-6">
        {mode === 'url' ? (
          <div>
            <label className="mb-1.5 block text-sm font-medium">GitHub URL</label>
            <input
              type="text"
              value={githubUrl}
              onChange={(e) => setGithubUrl(e.target.value)}
              placeholder="https://github.com/owner/repo"
              className="h-9 w-full rounded-md border border-input bg-background px-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
              onKeyDown={(e) => e.key === 'Enter' && handleConnect()}
            />
          </div>
        ) : (
          <div>
            <label className="mb-1.5 block text-sm font-medium">Select Repository</label>
            <div className="relative mb-3">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <input
                type="text"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Search repositories..."
                className="h-9 w-full rounded-md border border-input bg-background pl-10 pr-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
              />
            </div>

            {loadingRepos ? (
              <div className="flex items-center justify-center py-8 text-sm text-muted-foreground">
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Loading your repositories...
              </div>
            ) : reposError ? (
              <div className="py-4 text-center text-sm text-destructive">
                {reposError}
              </div>
            ) : (
              <div className="max-h-[280px] space-y-1 overflow-y-auto">
                {filteredRepos.map((repo) => (
                  <button
                    key={repo.nameWithOwner}
                    onClick={() => setSelectedRepo(repo)}
                    className={`flex w-full items-center gap-3 rounded-md px-3 py-2 text-left text-sm transition-colors ${
                      selectedRepo?.nameWithOwner === repo.nameWithOwner
                        ? 'bg-primary/10 text-primary'
                        : 'hover:bg-muted'
                    }`}
                  >
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="font-medium truncate">{repo.nameWithOwner}</span>
                        {repo.isPrivate ? (
                          <Lock className="h-3 w-3 shrink-0 text-muted-foreground" />
                        ) : (
                          <Globe className="h-3 w-3 shrink-0 text-muted-foreground" />
                        )}
                      </div>
                      {repo.description && (
                        <p className="mt-0.5 truncate text-xs text-muted-foreground">
                          {repo.description}
                        </p>
                      )}
                    </div>
                    {selectedRepo?.nameWithOwner === repo.nameWithOwner && (
                      <Check className="h-4 w-4 shrink-0 text-primary" />
                    )}
                  </button>
                ))}
                {filteredRepos.length === 0 && (
                  <p className="py-4 text-center text-xs text-muted-foreground">
                    No repositories found
                  </p>
                )}
              </div>
            )}
          </div>
        )}

        <div>
          <label className="mb-1.5 block text-sm font-medium">Clone To</label>
          <div className="relative">
            <FolderOpen className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <input
              type="text"
              value={parentDir}
              onChange={(e) => setParentDir(e.target.value)}
              className="h-9 w-full rounded-md border border-input bg-background pl-10 pr-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>
          {repoName && (
            <p className="mt-1.5 text-xs text-muted-foreground">
              Will clone to: <code>{parentDir}/{repoName}</code>
            </p>
          )}
        </div>

        {error && (
          <p className="text-sm text-destructive">{error}</p>
        )}

        <div className="flex justify-end gap-2">
          <Button variant="outline" size="sm" onClick={onBack}>
            Cancel
          </Button>
          <Button
            size="sm"
            onClick={handleConnect}
            disabled={!effectiveUrl || connecting}
          >
            {connecting && <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" />}
            Clone & Connect
          </Button>
        </div>
      </div>
    </div>
  )
}

function DeleteProjectDialog({
  project,
  onCancel,
  onConfirm
}: {
  project: Project
  onCancel: () => void
  onConfirm: (projectId: string, deleteFiles: boolean, deleteRepo: boolean) => Promise<string | null>
}): React.JSX.Element {
  const [deleteFiles, setDeleteFiles] = useState(true)
  const [deleteRepo, setDeleteRepo] = useState(false)
  const [deleting, setDeleting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleConfirm = async (): Promise<void> => {
    setDeleting(true)
    setError(null)
    const err = await onConfirm(project.id, deleteFiles, deleteRepo)
    if (err) {
      setError(err)
      setDeleting(false)
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60">
      <div className="w-full max-w-md rounded-lg border bg-card p-6 shadow-lg">
        <div className="flex items-center gap-3 text-destructive">
          <AlertTriangle className="h-5 w-5" />
          <h2 className="text-lg font-semibold">Delete {project.name}?</h2>
        </div>

        <div className="mt-4 space-y-3">
          <label className="flex cursor-pointer items-center gap-3">
            <input
              type="checkbox"
              checked={deleteFiles}
              onChange={(e) => setDeleteFiles(e.target.checked)}
              className="h-4 w-4 rounded border-input accent-primary"
            />
            <div>
              <p className="text-sm font-medium">Delete local files</p>
              <p className="text-xs text-muted-foreground">{project.path}</p>
            </div>
          </label>

          {project.githubUrl && (
            <label className="flex cursor-pointer items-center gap-3">
              <input
                type="checkbox"
                checked={deleteRepo}
                onChange={(e) => setDeleteRepo(e.target.checked)}
                className="h-4 w-4 rounded border-input accent-primary"
              />
              <div>
                <p className="text-sm font-medium">Delete GitHub repository</p>
                <p className="text-xs text-muted-foreground">{project.githubUrl}</p>
              </div>
            </label>
          )}
        </div>

        {!deleteFiles && !deleteRepo && (
          <p className="mt-3 text-xs text-muted-foreground">
            This will only remove the project from ADE.
          </p>
        )}

        {error && (
          <p className="mt-3 text-sm text-destructive">{error}</p>
        )}

        <div className="mt-6 flex justify-end gap-2">
          <Button variant="outline" size="sm" onClick={onCancel} disabled={deleting}>
            Cancel
          </Button>
          <Button
            variant="destructive"
            size="sm"
            onClick={handleConfirm}
            disabled={deleting}
          >
            {deleting && <Loader2 className="mr-2 h-3.5 w-3.5 animate-spin" />}
            Delete
          </Button>
        </div>
      </div>
    </div>
  )
}

function ToggleSwitch({
  checked,
  onChange
}: {
  checked: boolean
  onChange: (value: boolean) => void
}): React.JSX.Element {
  return (
    <button
      role="switch"
      aria-checked={checked}
      onClick={() => onChange(!checked)}
      className={`relative inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full transition-colors ${
        checked ? 'bg-primary' : 'bg-input'
      }`}
    >
      <span
        className={`inline-block h-4 w-4 rounded-full bg-background shadow-sm transition-transform ${
          checked ? 'translate-x-[18px]' : 'translate-x-[2px]'
        }`}
      />
    </button>
  )
}
