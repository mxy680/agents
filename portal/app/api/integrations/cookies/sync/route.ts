import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { isAdmin } from "@/lib/admin"
import { encrypt } from "@/lib/crypto"

/**
 * POST /api/integrations/cookies/sync
 *
 * Called by the Chrome extension to push fresh cookies for a provider.
 * Body: { provider: "zillow", cookies: "name=value; name2=value2" }
 */
export async function POST(request: NextRequest) {
  const supabase = await createClient()
  const {
    data: { user },
  } = await supabase.auth.getUser()

  if (!user || !isAdmin(user.email)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  let provider: string
  let cookies: string
  try {
    const body = await request.json()
    provider = body.provider
    cookies = body.cookies
  } catch {
    return NextResponse.json({ error: "Invalid request body" }, { status: 400 })
  }

  if (!provider || !cookies) {
    return NextResponse.json(
      { error: "provider and cookies are required" },
      { status: 400 }
    )
  }

  // Allowlist of providers that support cookie sync
  const ALLOWED_PROVIDERS = ["zillow", "streeteasy"]
  if (!ALLOWED_PROVIDERS.includes(provider)) {
    return NextResponse.json(
      { error: `Unknown provider: ${provider}` },
      { status: 400 }
    )
  }

  try {
    const credentials = JSON.stringify({ all_cookies: cookies })
    const encrypted = encrypt(credentials)
    const credHex = `\\x${Buffer.from(encrypted).toString("hex")}`
    const now = new Date().toISOString()

    const admin = createAdminClient()

    const { data: existing } = await admin
      .from("user_integrations")
      .select("id")
      .eq("provider", provider)
      .eq("user_id", user.id)
      .eq("status", "active")
      .limit(1)
      .maybeSingle()

    if (existing) {
      await admin
        .from("user_integrations")
        .update({ credentials: credHex, updated_at: now })
        .eq("id", existing.id)
    } else {
      await admin.from("user_integrations").insert({
        user_id: user.id,
        provider,
        label: provider.charAt(0).toUpperCase() + provider.slice(1),
        status: "active",
        credentials: credHex,
        created_at: now,
        updated_at: now,
      })
    }

    return NextResponse.json({
      ok: true,
      provider,
      cookie_length: cookies.length,
      synced_at: now,
    })
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    console.error(`[cookies/sync] Error for ${provider}:`, message)
    return NextResponse.json({ ok: false, error: message }, { status: 500 })
  }
}
