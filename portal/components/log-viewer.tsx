"use client"

import { useEffect, useRef, useState } from "react"

interface Deliverable {
  [key: string]: string
}

interface LogViewerProps {
  runId: string
  initialLog?: string
  initialStatus?: string
  initialDeliverables?: Deliverable
  onDone?: (status: string, deliverables: Deliverable) => void
}

/**
 * Terminal-style log viewer that streams live output via SSE.
 * Connects to /api/jobs/[runId]/logs and appends lines as they arrive.
 * Color-codes phase headers (cyan), errors (red), success lines (green).
 */
export function LogViewer({
  runId,
  initialLog = "",
  initialStatus,
  initialDeliverables,
  onDone,
}: LogViewerProps) {
  const [log, setLog] = useState(initialLog)
  const [done, setDone] = useState(
    initialStatus === "completed" || initialStatus === "failed"
  )
  const bottomRef = useRef<HTMLDivElement>(null)
  const esRef = useRef<EventSource | null>(null)

  useEffect(() => {
    // If already terminal state, no need to connect
    if (done) return

    const es = new EventSource(`/api/jobs/${runId}/logs`)
    esRef.current = es

    es.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data) as {
          type: "log" | "done"
          content?: string
          status?: string
          deliverables?: Deliverable
        }

        if (msg.type === "log" && msg.content) {
          setLog((prev) => prev + msg.content)
        } else if (msg.type === "done") {
          setDone(true)
          es.close()
          if (onDone && msg.status) {
            onDone(msg.status, msg.deliverables ?? {})
          }
        }
      } catch {
        // Ignore parse errors
      }
    }

    es.onerror = () => {
      es.close()
    }

    return () => {
      es.close()
      esRef.current = null
    }
  }, [runId, done, onDone])

  // Auto-scroll to bottom when new content arrives
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [log])

  const lines = log.split("\n")

  return (
    <div
      style={{
        backgroundColor: "#1e1e1e",
        color: "#d4d4d4",
        fontFamily: '"JetBrains Mono", "Fira Code", "Cascadia Code", monospace',
        fontSize: "13px",
        lineHeight: "1.6",
        borderRadius: "6px",
        padding: "16px",
        overflowY: "auto",
        maxHeight: "600px",
        whiteSpace: "pre-wrap",
        wordBreak: "break-all",
      }}
    >
      {lines.map((line, i) => (
        <LogLine key={i} line={line} />
      ))}
      {!done && (
        <span
          style={{
            display: "inline-block",
            width: "8px",
            height: "14px",
            backgroundColor: "#d4d4d4",
            animation: "blink 1s step-end infinite",
            verticalAlign: "text-bottom",
            marginLeft: "2px",
          }}
        />
      )}
      <div ref={bottomRef} />
      <style>{`@keyframes blink { 0%, 100% { opacity: 1; } 50% { opacity: 0; } }`}</style>
    </div>
  )
}

function LogLine({ line }: { line: string }) {
  const color = getLineColor(line)
  return (
    <div style={{ color, minHeight: "1em" }}>
      {line || " "}
    </div>
  )
}

function getLineColor(line: string): string {
  const lower = line.toLowerCase()

  // Phase headers — cyan
  if (line.startsWith("━━━") || /^phase \d+:/i.test(line)) {
    return "#4ec9b0"
  }

  // Error lines — red
  if (
    lower.includes("error") ||
    lower.includes("exception") ||
    lower.includes("traceback") ||
    lower.includes("failed") ||
    lower.includes("fatal")
  ) {
    return "#f48771"
  }

  // Success / completion lines — green
  if (
    lower.includes("complete") ||
    lower.includes("success") ||
    lower.includes("uploaded") ||
    lower.includes("done") ||
    lower.includes("pipeline complete")
  ) {
    return "#6a9955"
  }

  // Warning lines — yellow
  if (lower.includes("warning") || lower.includes("warn")) {
    return "#dcdcaa"
  }

  // Default — light gray
  return "#d4d4d4"
}
