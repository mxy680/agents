/**
 * Types and NDJSON parser for Agent SDK streaming events.
 * Shared between the local container runner and the orchestrator path.
 */

export interface ChatSSEEvent {
  event: "delta" | "tool_start" | "tool_input" | "tool_result" | "session" | "result" | "error"
  data: string
}

export interface ToolIdRef {
  currentToolId: string | null
}

/**
 * Maps a single parsed NDJSON object from the Agent SDK to zero or more ChatSSEEvents.
 */
export function mapAgentEvent(
  raw: Record<string, unknown>,
  toolInputAccum: Record<string, string>,
  toolIdRef: ToolIdRef
): ChatSSEEvent[] {
  const rawType = raw.type as string | undefined

  // Agent SDK wraps stream events: {type: "stream_event", event: {type: "content_block_delta", ...}}
  // Unwrap to get the inner event for content parsing.
  let event: Record<string, unknown>
  let type: string | undefined
  if (rawType === "stream_event" && raw.event && typeof raw.event === "object") {
    event = raw.event as Record<string, unknown>
    type = event.type as string | undefined
  } else {
    event = raw
    type = rawType
  }

  switch (type) {
    case "content_block_start": {
      const block = event.content_block as Record<string, unknown> | undefined
      if (block?.type === "tool_use") {
        const id = block.id as string
        const name = block.name as string
        toolIdRef.currentToolId = id
        toolInputAccum[id] = ""
        return [{ event: "tool_start", data: JSON.stringify({ name, id }) }]
      }
      return []
    }

    case "content_block_delta": {
      const delta = event.delta as Record<string, unknown> | undefined
      if (!delta) return []

      if (delta.type === "text_delta") {
        const text = delta.text as string
        return [{ event: "delta", data: text }]
      }

      if (delta.type === "input_json_delta") {
        const partial = delta.partial_json as string
        if (toolIdRef.currentToolId) {
          toolInputAccum[toolIdRef.currentToolId] = (toolInputAccum[toolIdRef.currentToolId] ?? "") + partial
        }
        return [{ event: "tool_input", data: partial }]
      }

      return []
    }

    case "content_block_stop": {
      if (toolIdRef.currentToolId && toolInputAccum[toolIdRef.currentToolId] !== undefined) {
        const accumulated = toolInputAccum[toolIdRef.currentToolId]
        toolIdRef.currentToolId = null
        return [{ event: "tool_input", data: accumulated }]
      }
      return []
    }

    case "tool_use_summary": {
      const summary = event.summary as string | undefined
      return [{ event: "tool_result", data: JSON.stringify({ summary: summary ?? "" }) }]
    }

    // Agent SDK sends tool results as {type: "user", tool_use_result: {...}}
    case "user": {
      const toolResult = raw.tool_use_result as Record<string, unknown> | undefined
      if (!toolResult) return []
      const msg = raw.message as Record<string, unknown> | undefined
      const content = (msg?.content as Array<Record<string, unknown>>) ?? []
      const resultBlock = content.find((c) => c.type === "tool_result")
      if (!resultBlock) return []
      const toolUseId = resultBlock.tool_use_id as string | undefined
      const rawContent = resultBlock.content
      const resultContent = typeof rawContent === "string"
        ? rawContent
        : rawContent != null
          ? JSON.stringify(rawContent)
          : undefined
      const summary = resultContent
        ?? (toolResult.stdout as string | undefined)
        ?? JSON.stringify(toolResult)
      return [{
        event: "tool_result",
        data: JSON.stringify({ summary, toolUseId }),
      }]
    }

    case "result": {
      // result events come at top level from Agent SDK (not wrapped in stream_event)
      const sessionId = (raw.session_id ?? event.session_id) as string | undefined
      const events: ChatSSEEvent[] = []
      if (sessionId) {
        events.push({ event: "session", data: sessionId })
      }
      events.push({
        event: "result",
        data: JSON.stringify({
          sessionId,
          result: raw.result ?? event.result,
          costUsd: raw.total_cost_usd ?? event.total_cost_usd,
          numTurns: raw.num_turns ?? event.num_turns,
          stopReason: raw.stop_reason ?? event.stop_reason,
        }),
      })
      return events
    }

    default:
      return []
  }
}

/**
 * Parses a buffer of text containing NDJSON lines into ChatSSEEvents.
 * Returns the parsed events and any remaining incomplete line.
 */
export function parseNDJSON(
  text: string,
  lineBuffer: string,
  toolInputAccum: Record<string, string>,
  toolIdRef: ToolIdRef
): { events: ChatSSEEvent[]; remainingBuffer: string } {
  const combined = lineBuffer + text
  const lines = combined.split("\n")
  const remaining = lines.pop() ?? ""
  const events: ChatSSEEvent[] = []

  for (const line of lines) {
    const trimmed = line.trim()
    if (!trimmed) continue

    let parsed: Record<string, unknown>
    try {
      parsed = JSON.parse(trimmed)
    } catch {
      continue
    }

    const mapped = mapAgentEvent(parsed, toolInputAccum, toolIdRef)
    events.push(...mapped)
  }

  return { events, remainingBuffer: remaining }
}
