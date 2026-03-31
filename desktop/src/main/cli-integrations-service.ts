import { execFile } from 'child_process'
import { promisify } from 'util'
import { getEngagentEnv } from './engagent-config'
import { homedir } from 'os'
import { join } from 'path'

const execFileAsync = promisify(execFile)
const shellPath = process.env.SHELL || '/bin/zsh'

export interface CliIntegration {
  readonly id: string
  readonly name: string
  readonly description: string
  readonly command: string
  readonly testCommand: string
  readonly requiresCredentials: boolean
}

const INTEGRATIONS: CliIntegration[] = [
  // Google services (OAuth — token refresh works)
  { id: 'gmail', name: 'Gmail', description: 'Messages, drafts, labels', command: 'gmail', testCommand: 'gmail messages list --limit=1 --json', requiresCredentials: true },
  { id: 'calendar', name: 'Google Calendar', description: 'Events, calendars', command: 'calendar', testCommand: 'calendar events list --limit=1 --json', requiresCredentials: true },
  { id: 'sheets', name: 'Google Sheets', description: 'Spreadsheets, values', command: 'sheets', testCommand: 'sheets spreadsheets list --limit=1 --json', requiresCredentials: true },
  { id: 'drive', name: 'Google Drive', description: 'Files, folders', command: 'drive', testCommand: 'drive files list --limit=1 --json', requiresCredentials: true },
  { id: 'docs', name: 'Google Docs', description: 'Documents', command: 'docs', testCommand: 'drive files list --limit=1 --query="mimeType=\'application/vnd.google-apps.document\'" --json', requiresCredentials: true },
  // GitHub (needs GITHUB_CLIENT_ID/SECRET env vars for token refresh)
  { id: 'github', name: 'GitHub', description: 'Repos, issues, PRs, actions', command: 'github', testCommand: 'github repos list --limit=1 --json', requiresCredentials: true },
  // Session-based providers
  { id: 'instagram', name: 'Instagram', description: 'Media, stories, comments', command: 'ig', testCommand: 'ig media list --limit=1 --json', requiresCredentials: true },
  { id: 'linkedin', name: 'LinkedIn', description: 'Connections, messages, jobs', command: 'li', testCommand: 'li connections list --limit=1 --json', requiresCredentials: true },
  { id: 'x', name: 'X (Twitter)', description: 'Tweets, likes, DMs, lists', command: 'x', testCommand: 'x notifications all --json', requiresCredentials: true },
  // API key providers
  { id: 'vercel', name: 'Vercel', description: 'Deployments, projects', command: 'vercel', testCommand: 'vercel projects list --limit=1 --json', requiresCredentials: true },
  { id: 'cloudflare', name: 'Cloudflare', description: 'Zones, DNS, workers', command: 'cloudflare', testCommand: 'cloudflare zones list --json', requiresCredentials: true },
  { id: 'linear', name: 'Linear', description: 'Issues, projects, teams', command: 'linear', testCommand: 'linear teams list --json', requiresCredentials: true },
  { id: 'fly', name: 'Fly.io', description: 'Apps, regions, machines', command: 'fly', testCommand: 'fly regions list --json', requiresCredentials: true },
  // Services that need extra config — failures indicate real infra issues
  { id: 'imessage', name: 'iMessage', description: 'Messages via BlueBubbles (requires Mac + tunnel)', command: 'imessage', testCommand: 'imessage messages query --limit=1 --json', requiresCredentials: true },
  { id: 'supabase', name: 'Supabase', description: 'Databases, projects (re-auth in portal if expired)', command: 'sb', testCommand: 'sb projects list --json', requiresCredentials: true },
  { id: 'gcp', name: 'Google Cloud', description: 'Services, IAM, projects', command: 'gcp', testCommand: 'gcp services list --json', requiresCredentials: true },
  { id: 'framer', name: 'Framer', description: 'Pages, collections (connect in portal first)', command: 'framer', testCommand: 'framer collections list --json', requiresCredentials: true },
  // Public APIs (no credentials needed)
  { id: 'census', name: 'US Census', description: 'ACS 5-Year demographic data', command: 'census', testCommand: 'census --help', requiresCredentials: false },
  { id: 'citibike', name: 'Citi Bike', description: 'Station availability', command: 'citibike', testCommand: 'citibike --help', requiresCredentials: false },
  { id: 'places', name: 'Google Places', description: 'Business & place search', command: 'places', testCommand: 'places --help', requiresCredentials: false },
]

function getCliEnv(): Record<string, string> {
  return { ...process.env as Record<string, string>, ...getEngagentEnv() }
}

export function listCliIntegrations(): { integrations: CliIntegration[] } {
  return { integrations: INTEGRATIONS }
}

export async function testCliIntegration(
  _event: unknown,
  input: { integrationId: string }
): Promise<{ success: boolean; output: string; error: string | null }> {
  const integration = INTEGRATIONS.find((i) => i.id === input.integrationId)
  if (!integration) {
    return { success: false, output: '', error: 'Integration not found' }
  }

  const cliPath = join(homedir(), '.ade', 'bin', 'integrations')

  try {
    const { stdout, stderr } = await execFileAsync(
      shellPath,
      ['-l', '-c', `"${cliPath}" ${integration.testCommand} < /dev/null`],
      { env: getCliEnv(), timeout: 30000 }
    )

    const output = stdout.trim()
    const warnings = stderr.trim()

    return {
      success: true,
      output: output.slice(0, 1000) + (output.length > 1000 ? '...' : ''),
      error: warnings || null
    }
  } catch (err) {
    const message = err instanceof Error ? err.message : 'Unknown error'
    return { success: false, output: '', error: message.slice(0, 500) }
  }
}
