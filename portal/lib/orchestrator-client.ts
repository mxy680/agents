/**
 * Client for the orchestrator REST API.
 * The orchestrator manages agent container lifecycle on the cloud server.
 */

export interface AgentInstance {
  id: string
  user_id: string
  template_id: string
  status: string
  k8s_pod_name: string
  k8s_namespace: string
  config_overrides: Record<string, unknown>
  error_message?: string
  started_at?: string
  stopped_at?: string
  created_at: string
  updated_at: string
}

function getBaseUrl(): string {
  const baseUrl = process.env.ORCHESTRATOR_URL
  if (!baseUrl) {
    throw new Error("ORCHESTRATOR_URL is not configured")
  }
  return baseUrl.replace(/\/+$/, "")
}

function authHeaders(_authToken?: string): Record<string, string> {
  // Use static API key for internal admin tool (more reliable than Supabase JWT)
  const apiKey = process.env.ORCHESTRATOR_API_KEY
  if (apiKey) {
    return { Authorization: `Bearer ${apiKey}` }
  }
  // Fallback to passed token (Supabase JWT)
  return _authToken ? { Authorization: `Bearer ${_authToken}` } : {}
}

/**
 * Deploy an agent container.
 * Returns the created AgentInstance (status will be "creating").
 */
export async function deployAgent(
  templateId: string,
  configOverrides?: Record<string, unknown>,
  authToken?: string
): Promise<AgentInstance> {
  const baseUrl = getBaseUrl()
  const res = await fetch(`${baseUrl}/api/v1/agents/deploy`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      ...authHeaders(authToken),
    },
    body: JSON.stringify({
      template_id: templateId,
      config_overrides: configOverrides ?? {},
    }),
  })

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText })) as { error?: string }
    throw new Error(`Orchestrator deploy failed: ${err.error ?? res.statusText}`)
  }

  return res.json() as Promise<AgentInstance>
}

/**
 * Stream logs from a running agent instance.
 * The orchestrator sends SSE-formatted data: each chunk is `data: <raw stdout>\n\n`.
 * Returns the raw Response body as a ReadableStream.
 */
export async function streamAgentLogs(
  instanceId: string,
  authToken?: string,
  signal?: AbortSignal
): Promise<ReadableStream<Uint8Array>> {
  const baseUrl = getBaseUrl()
  const res = await fetch(`${baseUrl}/api/v1/agents/${instanceId}/logs`, {
    headers: {
      ...authHeaders(authToken),
    },
    signal,
  })

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText })) as { error?: string }
    throw new Error(`Orchestrator logs failed: ${err.error ?? res.statusText}`)
  }

  if (!res.body) {
    throw new Error("No response body from orchestrator logs endpoint")
  }

  return res.body
}

/**
 * Stop a running agent instance.
 */
export async function stopAgent(
  instanceId: string,
  authToken?: string
): Promise<void> {
  const baseUrl = getBaseUrl()
  await fetch(`${baseUrl}/api/v1/agents/${instanceId}/stop`, {
    method: "POST",
    headers: {
      ...authHeaders(authToken),
    },
  })
}

/**
 * Get agent instance status.
 */
export async function getAgentStatus(
  instanceId: string,
  authToken?: string
): Promise<AgentInstance> {
  const baseUrl = getBaseUrl()
  const res = await fetch(`${baseUrl}/api/v1/agents/${instanceId}`, {
    headers: {
      ...authHeaders(authToken),
    },
  })

  if (!res.ok) {
    throw new Error(`Failed to get agent status: ${res.statusText}`)
  }

  return res.json() as Promise<AgentInstance>
}
