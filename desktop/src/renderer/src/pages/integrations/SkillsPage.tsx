import { useCallback, useEffect, useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Loader2, Zap, Globe, FolderGit2, Search, ChevronDown, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'

interface Skill {
  readonly name: string
  readonly description: string
  readonly origin: string
  readonly scope: 'global' | 'project'
  readonly path: string
}

interface Project {
  readonly id: string
  readonly name: string
  readonly path: string
}

export function SkillsPage(): React.JSX.Element {
  const [skills, setSkills] = useState<readonly Skill[]>([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState('')
  const [projects, setProjects] = useState<readonly Project[]>([])
  const [selectedProjectId, setSelectedProjectId] = useState<string>('all')
  const [dropdownOpen, setDropdownOpen] = useState(false)

  // Load projects
  useEffect(() => {
    window.api.projects.list().then((result) => {
      setProjects(result.projects)
    })
  }, [])

  const loadSkills = useCallback(() => {
    setLoading(true)
    const selectedProject = projects.find((p) => p.id === selectedProjectId)
    window.api.skills
      .list(selectedProject ? { projectPath: selectedProject.path } : undefined)
      .then((result) => {
        setSkills(result.skills)
        setLoading(false)
      })
  }, [selectedProjectId, projects])

  useEffect(() => {
    loadSkills()
  }, [loadSkills])

  const handleRemove = async (skill: Skill): Promise<void> => {
    await window.api.skills.remove({ skillPath: skill.path })
    loadSkills()
  }

  const globalSkills = skills.filter((s) => s.scope === 'global')
  const projectSkills = skills.filter((s) => s.scope === 'project')

  const filterSkills = (list: readonly Skill[]): readonly Skill[] => {
    if (!search.trim()) return list
    const q = search.toLowerCase()
    return list.filter(
      (s) =>
        s.name.toLowerCase().includes(q) ||
        s.description.toLowerCase().includes(q)
    )
  }

  const filteredGlobal = filterSkills(globalSkills)
  const filteredProject = filterSkills(projectSkills)

  return (
    <div className="mx-auto max-w-2xl space-y-8 p-8">
      {/* Header */}
      <div>
        <div className="flex items-center gap-3">
          <div className="flex size-12 items-center justify-center rounded-lg bg-secondary">
            <Zap className="size-6" />
          </div>
          <div>
            <h1 className="text-xl font-semibold">Skills</h1>
            <p className="text-sm text-muted-foreground">
              Browse configured skills for Claude Code
            </p>
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="flex gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search skills..."
            className="h-9 w-full rounded-md border border-input bg-background pl-10 pr-3 text-sm placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </div>

        {/* Project filter */}
        <div className="relative">
          <button
            onClick={() => setDropdownOpen((v) => !v)}
            className="flex h-9 items-center gap-2 rounded-md border border-input bg-background px-3 text-sm hover:bg-accent"
          >
            <FolderGit2 className="size-3.5 text-muted-foreground" />
            <span className="max-w-[120px] truncate">
              {selectedProjectId === 'all'
                ? 'All projects'
                : projects.find((p) => p.id === selectedProjectId)?.name ?? 'Select'}
            </span>
            <ChevronDown className="size-3.5 text-muted-foreground" />
          </button>

          {dropdownOpen && (
            <div className="absolute right-0 top-10 z-50 w-48 rounded-md border bg-card py-1 shadow-lg">
              <button
                onClick={() => { setSelectedProjectId('all'); setDropdownOpen(false) }}
                className={cn(
                  'flex w-full items-center px-3 py-1.5 text-sm transition-colors hover:bg-accent',
                  selectedProjectId === 'all' && 'text-primary font-medium'
                )}
              >
                All projects
              </button>
              {projects.map((project) => (
                <button
                  key={project.id}
                  onClick={() => { setSelectedProjectId(project.id); setDropdownOpen(false) }}
                  className={cn(
                    'flex w-full items-center px-3 py-1.5 text-sm transition-colors hover:bg-accent',
                    selectedProjectId === project.id && 'text-primary font-medium'
                  )}
                >
                  {project.name}
                </button>
              ))}
            </div>
          )}
        </div>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-12 text-sm text-muted-foreground">
          <Loader2 className="mr-2 size-4 animate-spin" />
          Loading skills...
        </div>
      ) : (
        <>
          {/* Project Skills */}
          {filteredProject.length > 0 && (
            <div>
              <div className="mb-3 flex items-center gap-2 text-sm font-medium">
                <FolderGit2 className="size-4 text-muted-foreground" />
                Project Skills
                <Badge variant="secondary" className="text-[10px]">{filteredProject.length}</Badge>
              </div>
              <div className="space-y-2">
                {filteredProject.map((skill) => (
                  <SkillCard key={skill.path} skill={skill} onRemove={() => handleRemove(skill)} />
                ))}
              </div>
            </div>
          )}

          {/* Global Skills */}
          {filteredGlobal.length > 0 && (
            <div>
              <div className="mb-3 flex items-center gap-2 text-sm font-medium">
                <Globe className="size-4 text-muted-foreground" />
                Global Skills
                <Badge variant="secondary" className="text-[10px]">{filteredGlobal.length}</Badge>
              </div>
              <div className="space-y-2">
                {filteredGlobal.map((skill) => (
                  <SkillCard key={skill.path} skill={skill} onRemove={() => handleRemove(skill)} />
                ))}
              </div>
            </div>
          )}

          {filteredGlobal.length === 0 && filteredProject.length === 0 && (
            <div className="py-8 text-center text-sm text-muted-foreground">
              {search ? 'No skills match your search.' : 'No skills configured.'}
            </div>
          )}
        </>
      )}
    </div>
  )
}

function SkillCard({ skill, onRemove }: { skill: Skill; onRemove: () => void }): React.JSX.Element {
  return (
    <div className="group rounded-lg border bg-card p-4">
      <div className="flex items-start justify-between">
        <div className="flex items-center gap-2">
          <Zap className="size-4 text-primary" />
          <span className="text-sm font-medium">{skill.name}</span>
          {skill.origin && (
            <Badge variant="secondary" className="text-[10px]">{skill.origin}</Badge>
          )}
        </div>
        <button
          onClick={onRemove}
          className="rounded-md p-1 text-muted-foreground opacity-0 transition-all hover:text-destructive group-hover:opacity-100"
          title="Remove skill"
        >
          <Trash2 className="size-3.5" />
        </button>
      </div>
      {skill.description && (
        <p className="mt-1.5 text-xs text-muted-foreground leading-relaxed">
          {skill.description}
        </p>
      )}
    </div>
  )
}
