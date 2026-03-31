import { readdir, readFile, rm } from 'fs/promises'
import { join } from 'path'
import { homedir } from 'os'

export interface Skill {
  readonly name: string
  readonly description: string
  readonly origin: string
  readonly scope: 'global' | 'project'
  readonly path: string
}

interface SkillsResult {
  readonly skills: readonly Skill[]
  readonly error: string | null
}

async function parseSkillMd(skillDir: string, scope: 'global' | 'project'): Promise<Skill | null> {
  try {
    const mdPath = join(skillDir, 'SKILL.md')
    const content = await readFile(mdPath, 'utf-8')

    // Parse YAML frontmatter
    const match = content.match(/^---\n([\s\S]*?)\n---/)
    if (!match) return null

    const frontmatter = match[1]
    const name = frontmatter.match(/^name:\s*(.+)$/m)?.[1]?.trim() ?? ''
    const description = frontmatter.match(/^description:\s*(.+)$/m)?.[1]?.trim() ?? ''
    const origin = frontmatter.match(/^origin:\s*(.+)$/m)?.[1]?.trim() ?? ''

    if (!name) return null

    return { name, description, origin, scope, path: skillDir }
  } catch {
    return null
  }
}

async function scanSkillsDir(dir: string, scope: 'global' | 'project'): Promise<Skill[]> {
  try {
    const entries = await readdir(dir, { withFileTypes: true })
    const skills: Skill[] = []

    for (const entry of entries) {
      if (!entry.isDirectory()) continue
      const skill = await parseSkillMd(join(dir, entry.name), scope)
      if (skill) skills.push(skill)
    }

    return skills
  } catch {
    return []
  }
}

export async function listSkills(
  _event: unknown,
  input?: { projectPath?: string }
): Promise<SkillsResult> {
  try {
    const globalDir = join(homedir(), '.claude', 'skills')
    const globalSkills = await scanSkillsDir(globalDir, 'global')

    let projectSkills: Skill[] = []
    if (input?.projectPath) {
      const projectDir = join(input.projectPath, '.claude', 'skills')
      projectSkills = await scanSkillsDir(projectDir, 'project')
    }

    return {
      skills: [...projectSkills, ...globalSkills],
      error: null
    }
  } catch (err) {
    return {
      skills: [],
      error: err instanceof Error ? err.message : 'Failed to list skills'
    }
  }
}

export async function removeSkill(
  _event: unknown,
  input: { skillPath: string }
): Promise<{ error: string | null }> {
  try {
    await rm(input.skillPath, { recursive: true, force: true })
    return { error: null }
  } catch (err) {
    return { error: err instanceof Error ? err.message : 'Failed to remove skill' }
  }
}
