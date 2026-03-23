/**
 * Maps Agent SDK tool calls to human-readable labels and icon names.
 * The agents primarily use the Bash tool to run `integrations <provider> <resource> <action>`.
 */

export interface ToolDisplay {
  label: string
  icon: string // Tabler icon name (without Icon prefix)
  provider?: string // e.g. "gmail", "github"
  command?: string // the raw command if it's a bash tool
}

const PROVIDER_ICONS: Record<string, string> = {
  gmail: "IconMail",
  google: "IconBrandGoogle",
  sheets: "IconTable",
  calendar: "IconCalendar",
  drive: "IconFolder",
  github: "IconBrandGithub",
  instagram: "IconBrandInstagram",
  linkedin: "IconBrandLinkedin",
  x: "IconBrandX",
  canvas: "IconSchool",
  supabase: "IconBrandSupabase",
  imessage: "IconMessage",
  framer: "IconLayout",
  places: "IconMapPin",
}

const PROVIDER_LABELS: Record<string, string> = {
  gmail: "Gmail",
  sheets: "Google Sheets",
  calendar: "Google Calendar",
  drive: "Google Drive",
  github: "GitHub",
  instagram: "Instagram",
  linkedin: "LinkedIn",
  x: "X (Twitter)",
  canvas: "Canvas LMS",
  supabase: "Supabase",
  imessage: "iMessage",
  framer: "Framer",
  places: "Google Places",
}

const ACTION_LABELS: Record<string, string> = {
  list: "Listing",
  get: "Fetching",
  create: "Creating",
  update: "Updating",
  delete: "Deleting",
  send: "Sending",
  search: "Searching",
  deploy: "Deploying",
  upload: "Uploading",
  download: "Downloading",
}

/**
 * Parse a tool call into a human-readable display.
 *
 * The Agent SDK sends tool names like "Bash", "Read", "Write", "Glob", etc.
 * For Bash tools, the input JSON contains a "command" field.
 * We parse `integrations <provider> <resource> <action>` from the command.
 */
export function parseToolDisplay(toolName: string, toolInput: string): ToolDisplay {
  // Non-bash tools
  if (toolName === "Read") {
    return { label: "Reading file", icon: "IconFileText" }
  }
  if (toolName === "Write") {
    return { label: "Writing file", icon: "IconFileText" }
  }
  if (toolName === "Edit") {
    return { label: "Editing file", icon: "IconEdit" }
  }
  if (toolName === "Glob") {
    return { label: "Searching files", icon: "IconSearch" }
  }
  if (toolName === "Grep") {
    return { label: "Searching code", icon: "IconSearch" }
  }
  if (toolName === "WebFetch") {
    return { label: "Fetching web page", icon: "IconWorld" }
  }
  if (toolName === "WebSearch") {
    return { label: "Searching the web", icon: "IconWorld" }
  }

  // Bash tool — parse the command
  if (toolName === "Bash" || toolName === "bash") {
    let command = ""
    try {
      const parsed = JSON.parse(toolInput)
      command = parsed.command || parsed.input || ""
    } catch {
      // toolInput might be the raw command string during streaming
      command = toolInput
    }

    // Try to match `integrations <provider> <resource> <action>`
    const match = command.match(/integrations\s+(\w+)\s+(\w+)\s+(\w+)/)
    if (match) {
      const [, provider, resource, action] = match
      const providerLabel = PROVIDER_LABELS[provider] || provider
      const actionLabel = ACTION_LABELS[action] || action
      const icon = PROVIDER_ICONS[provider] || "IconTerminal"

      return {
        label: `${providerLabel}: ${actionLabel} ${resource}`,
        icon,
        provider,
        command,
      }
    }

    // Fallback for non-integrations bash commands
    if (command.length > 60) {
      return { label: "Running command", icon: "IconTerminal", command }
    }
    return { label: "Running command", icon: "IconTerminal", command }
  }

  // Unknown tool
  return { label: toolName, icon: "IconTool" }
}

/**
 * Get a short result summary from tool output.
 * Tries to extract meaningful info from JSON results.
 */
export function parseToolResult(result: string): string {
  if (!result) return "Done"

  // Try to parse as JSON array and count items
  try {
    const parsed = JSON.parse(result)
    if (Array.isArray(parsed)) {
      return `Found ${parsed.length} ${parsed.length === 1 ? "item" : "items"}`
    }
    if (typeof parsed === "object" && parsed !== null) {
      if (parsed.error) return `Error: ${parsed.error}`
      if (parsed.id) return `Success (${parsed.id})`
      return "Done"
    }
  } catch {
    // Not JSON
  }

  // Truncate long text results
  if (result.length > 100) {
    return result.slice(0, 97) + "..."
  }
  return result
}
