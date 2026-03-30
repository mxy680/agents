import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { isAdmin } from "@/lib/admin"
import { credentialsToEnv, decryptAndRefresh } from "@/lib/credentials"
import { execFile } from "child_process"
import { promisify } from "util"
import path from "path"
import { execFileSync } from "child_process"

const execFileAsync = promisify(execFile)

/** Resolve the integrations binary: env var → PATH → local dev fallback */
function resolveIntegrationsBin(): string {
  if (process.env.INTEGRATIONS_BIN_PATH) {
    return process.env.INTEGRATIONS_BIN_PATH
  }
  // Check if `integrations` is on PATH (production Docker image)
  try {
    const resolved = execFileSync("which", ["integrations"], {
      encoding: "utf-8",
    }).trim()
    if (resolved) return resolved
  } catch {
    // not on PATH
  }
  // Local dev fallback: ../bin/integrations relative to portal/
  return path.resolve(process.cwd(), "..", "bin", "integrations")
}

/**
 * Simple read-only test command for each provider.
 * Each returns quickly and verifies credentials are valid.
 */
/**
 * Simple read-only test command for each provider.
 * Each returns quickly and verifies credentials are valid.
 * null = credentials-only check (no CLI command available locally).
 */
const TEST_COMMANDS: Record<string, string[] | null> = {
  google: ["gmail", "messages", "list", "--limit=1", "--json"],
  github: ["github", "repos", "list", "--limit=1", "--json"],
  instagram: ["instagram", "media", "list", "--limit=1", "--json"],
  linkedin: ["linkedin", "profile", "me", "--json"],
  x: ["x", "users", "get", "--username=elonmusk", "--json"],
  framer: null,  // Requires Node.js bridge — verify credentials only
  supabase: ["supabase", "projects", "list", "--json"],
  bluebubbles: ["imessage", "server", "info", "--json"],
  canvas: ["canvas", "users", "me", "--json"],
  zillow: ["zillow", "properties", "search", "--location=Denver, CO", "--limit=1", "--json"],
  vercel: ["vercel", "teams", "list", "--json"],
  cloudflare: ["cloudflare", "zones", "list", "--json"],
  linear: ["linear", "users", "me", "--json"],
  fly: ["fly", "apps", "list", "--org=personal", "--json"],
  "gcp-console": ["gcp-console", "oauth", "list", "--project-number=58889913836", "--json"],
  gcp: ["gcp", "projects", "get", "--project=engagent", "--json"],
}

export async function POST(request: NextRequest) {
  const supabase = await createClient()
  const {
    data: { user },
  } = await supabase.auth.getUser()

  if (!user || !isAdmin(user.email)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  const body = await request.json().catch(() => ({}))
  const integrationId = typeof body.id === "string" ? body.id : ""

  if (!integrationId) {
    return NextResponse.json({ error: "Integration ID required" }, { status: 400 })
  }

  // Fetch the integration
  const admin = createAdminClient()
  const { data: integration, error: fetchErr } = await admin
    .from("user_integrations")
    .select("id, provider, credentials, status")
    .eq("id", integrationId)
    .eq("user_id", user.id)
    .single()

  if (fetchErr || !integration) {
    return NextResponse.json({ error: "Integration not found" }, { status: 404 })
  }

  if (!(integration.provider in TEST_COMMANDS)) {
    return NextResponse.json({ error: `No test for provider: ${integration.provider}` }, { status: 400 })
  }

  // Decrypt credentials, refresh if expired, then map to env vars
  let env: Record<string, string> = {}
  try {
    const credJson = await decryptAndRefresh(integration)
    env = credentialsToEnv(integration.provider, credJson)
  } catch {
    return NextResponse.json({ ok: false, error: "Failed to decrypt credentials" })
  }

  // If no CLI command, just verify credentials decrypted with non-empty values
  const testArgs = TEST_COMMANDS[integration.provider]
  if (testArgs === null) {
    const hasValues = Object.values(env).some((v) => v.length > 0)
    return NextResponse.json({
      ok: hasValues,
      provider: integration.provider,
      ...(!hasValues && { error: "Credentials are empty" }),
    })
  }

  const binPath = resolveIntegrationsBin()

  try {
    await execFileAsync(binPath, testArgs, {
      env: { ...process.env, ...env, PATH: process.env.PATH },
      timeout: 15_000,
    })

    return NextResponse.json({ ok: true, provider: integration.provider })
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : String(err)
    const stderr = (err as { stderr?: string }).stderr || ""
    return NextResponse.json({
      ok: false,
      provider: integration.provider,
      error: stderr.slice(0, 200) || message.slice(0, 200),
    })
  }
}
