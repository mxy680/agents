"use client"

import * as React from "react"
import {
  IconLoader2,
  IconCheck,
  IconChevronDown,
  IconChevronRight,
  IconTerminal,
  IconMail,
  IconBrandGithub,
  IconBrandInstagram,
  IconBrandLinkedin,
  IconBrandX,
  IconSchool,
  IconMessage,
  IconLayout,
  IconSearch,
  IconFileText,
  IconEdit,
  IconWorld,
  IconTool,
  IconTable,
  IconCalendar,
  IconFolder,
  IconMapPin,
  IconBrandGoogle,
} from "@tabler/icons-react"
import { parseToolDisplay, parseToolResult } from "@/lib/tool-icons"

// Map icon string names to actual components
const ICON_MAP: Record<string, React.ComponentType<{ className?: string }>> = {
  IconTerminal,
  IconMail,
  IconBrandGithub,
  IconBrandInstagram,
  IconBrandLinkedin,
  IconBrandX,
  IconSchool,
  IconMessage,
  IconLayout,
  IconSearch,
  IconFileText,
  IconEdit,
  IconWorld,
  IconTool,
  IconTable,
  IconCalendar,
  IconFolder,
  IconMapPin,
  IconBrandGoogle,
}

// Provider-specific border colors
const PROVIDER_COLORS: Record<string, string> = {
  gmail: "border-l-red-500",
  github: "border-l-gray-400",
  instagram: "border-l-pink-500",
  linkedin: "border-l-blue-600",
  x: "border-l-gray-300",
  canvas: "border-l-red-600",
  supabase: "border-l-green-500",
  imessage: "border-l-blue-500",
  framer: "border-l-purple-500",
}

interface ToolCardProps {
  name: string
  id: string
  input: string
  result?: string
  isStreaming: boolean // is the parent message still streaming?
}

export function ToolCard({ name, id: _id, input, result, isStreaming }: ToolCardProps) {
  const display = parseToolDisplay(name, input)
  const IconComponent = ICON_MAP[display.icon] || IconTool
  const borderColor = display.provider
    ? (PROVIDER_COLORS[display.provider] ?? "border-l-muted-foreground")
    : "border-l-muted-foreground"

  const isDone = !!result
  const isRunning = isStreaming && !isDone

  // Auto-expand while running, auto-collapse when done
  const [expanded, setExpanded] = React.useState(false)

  React.useEffect(() => {
    if (isRunning) {
      setExpanded(true)
    } else if (isDone) {
      // Auto-collapse after a brief delay so user sees the result
      const timer = setTimeout(() => setExpanded(false), 800)
      return () => clearTimeout(timer)
    }
  }, [isRunning, isDone])

  const resultSummary = isDone ? parseToolResult(result) : null

  return (
    <div
      className={`border-l-2 ${borderColor} bg-muted/30 rounded-r-md overflow-hidden transition-all cursor-pointer`}
      onClick={() => setExpanded(!expanded)}
    >
      {/* Header — always visible */}
      <div className="flex items-center gap-2 px-3 py-2">
        {isRunning ? (
          <IconLoader2 className="size-4 animate-spin text-muted-foreground" />
        ) : isDone ? (
          <IconCheck className="size-4 text-green-500" />
        ) : (
          <IconComponent className="size-4 text-muted-foreground" />
        )}
        <span className="text-sm font-medium flex-1">{display.label}</span>
        {isDone && resultSummary && !expanded && (
          <span className="text-xs text-muted-foreground">{resultSummary}</span>
        )}
        {expanded ? (
          <IconChevronDown className="size-3.5 text-muted-foreground" />
        ) : (
          <IconChevronRight className="size-3.5 text-muted-foreground" />
        )}
      </div>

      {/* Expanded content */}
      {expanded && (
        <div className="px-3 pb-3 space-y-2">
          {/* Raw command */}
          {display.command && (
            <pre className="text-xs bg-background/50 rounded p-2 overflow-x-auto font-mono text-muted-foreground">
              {display.command}
            </pre>
          )}
          {/* Result */}
          {isDone && result && (
            <div className="text-xs">
              <div className="text-muted-foreground mb-1 font-medium">Result:</div>
              <pre className="bg-background/50 rounded p-2 overflow-x-auto font-mono whitespace-pre-wrap max-h-48 overflow-y-auto">
                {result.length > 2000 ? result.slice(0, 2000) + "\n... (truncated)" : result}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
