import { NextRequest, NextResponse } from "next/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { verifyExtensionToken } from "@/lib/extension-token"
import { encrypt } from "@/lib/crypto"

/**
 * Map of provider → cookie keys that must be present for the integration to be
 * usable. Optional cookies (mid, ig_did, bcookie, etc.) are stored if provided
 * but not required for validation.
 */
const PROVIDER_REQUIRED_COOKIES: Record<string, string[]> = {
  instagram: ["session_id"],
  linkedin: ["li_at", "jsessionid"],
  x: ["auth_token", "csrf_token"],
}

// Chrome extensions send Origin: chrome-extension://<id>
// Only allow our known extension ID and portal origins.
const ALLOWED_ORIGINS = [
  "chrome-extension://pkkpglobhebcecahhomkiniapgdpfico",
  "http://localhost:3000",
  "https://app.emdash.io",
  "https://agents.emdash.io",
]

function corsHeaders(request: NextRequest) {
  const origin = request.headers.get("origin") ?? ""
  const allowedOrigin = ALLOWED_ORIGINS.includes(origin) ? origin : ALLOWED_ORIGINS[0]
  return {
    "Access-Control-Allow-Origin": allowedOrigin,
    "Access-Control-Allow-Methods": "POST, OPTIONS",
    "Access-Control-Allow-Headers": "Content-Type, Authorization",
  }
}

/**
 * OPTIONS /api/integrations/extension/cookies
 * CORS preflight for Chrome extension service worker requests.
 */
export async function OPTIONS(request: NextRequest) {
  return new NextResponse(null, { status: 204, headers: corsHeaders(request) })
}

/**
 * POST /api/integrations/extension/cookies
 *
 * Receives cookies from the Chrome extension, encrypts them, and upserts an
 * integration record for the authenticated user.
 *
 * Auth: Bearer token (extension token from /api/integrations/extension/token).
 * CSRF origin check is intentionally skipped — the request originates from the
 * extension service worker, not the portal page.
 *
 * Body: { provider: string, cookies: Record<string, string>, label?: string }
 */
export async function POST(request: NextRequest) {
  const headers = corsHeaders(request)

  function corsJson(data: unknown, init?: { status?: number }) {
    return NextResponse.json(data, { ...init, headers })
  }

  // Auth via Bearer token
  const authHeader = request.headers.get("authorization")
  if (!authHeader?.startsWith("Bearer ")) {
    return corsJson({ error: "Missing authorization header" }, { status: 401 })
  }

  let userId: string
  try {
    const result = verifyExtensionToken(authHeader.slice(7))
    userId = result.userId
  } catch {
    return corsJson({ error: "Invalid or expired token" }, { status: 401 })
  }

  const body = await request.json().catch(() => null)
  if (
    !body ||
    typeof body.provider !== "string" ||
    typeof body.cookies !== "object" ||
    body.cookies === null ||
    Array.isArray(body.cookies)
  ) {
    return corsJson({ error: "Invalid request body" }, { status: 400 })
  }

  const { provider, cookies, label } = body as {
    provider: string
    cookies: Record<string, string>
    label?: unknown
  }

  const requiredCookies = PROVIDER_REQUIRED_COOKIES[provider]
  if (!requiredCookies) {
    return corsJson({ error: `Unknown provider: ${provider}` }, { status: 400 })
  }

  // Validate required cookies are present and non-empty
  for (const key of requiredCookies) {
    if (typeof cookies[key] !== "string" || !cookies[key]) {
      return corsJson({ error: `Missing required cookie: ${key}` }, { status: 400 })
    }
  }

  // Encrypt cookie payload
  const payload = JSON.stringify(cookies)
  const encrypted = encrypt(payload)

  const accountLabel =
    typeof label === "string" && label.trim()
      ? label.trim().slice(0, 100)
      : `${provider.charAt(0).toUpperCase()}${provider.slice(1)} Account`

  const admin = createAdminClient()
  const { error } = await admin.from("user_integrations").upsert(
    {
      user_id: userId,
      provider,
      label: accountLabel,
      status: "active",
      credentials: `\\x${encrypted.toString("hex")}`,
      updated_at: new Date().toISOString(),
    },
    { onConflict: "user_id,provider,label" }
  )

  if (error) {
    console.error("[extension/cookies] DB error:", error)
    return corsJson({ error: "Failed to save credentials" }, { status: 500 })
  }

  return corsJson({ success: true, provider })
}
