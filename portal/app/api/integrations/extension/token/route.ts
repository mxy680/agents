import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { checkOrigin } from "@/lib/csrf"
import { checkRateLimit } from "@/lib/rate-limit"
import { createExtensionToken } from "@/lib/extension-token"

/**
 * POST /api/integrations/extension/token
 *
 * Generates a long-lived Bearer token for the Chrome extension to use when
 * calling /api/integrations/extension/cookies. The token is HMAC-signed and
 * expires after 30 days. Rate-limited to 5 tokens per user per hour to prevent
 * abuse.
 *
 * Must be called from the portal page (Origin header checked via checkOrigin).
 */
export async function POST(request: NextRequest) {
  const csrfError = checkOrigin(request)
  if (csrfError) return csrfError

  const supabase = await createClient()
  const {
    data: { user },
  } = await supabase.auth.getUser()

  if (!user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  if (!checkRateLimit(`ext-token:${user.id}`, 30, 3_600_000)) {
    return NextResponse.json({ error: "Too many requests" }, { status: 429 })
  }

  const token = createExtensionToken(user.id)
  return NextResponse.json({ token })
}
