import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { isAdmin } from "@/lib/admin"
import { checkOrigin } from "@/lib/csrf"
import { encrypt } from "@/lib/crypto"
import {
  mapInstagramCookies,
  mapLinkedinCookies,
  mapXCookies,
  mapCanvasCookies,
  mapZillowCookies,
  mapStreetEasyCookies,
  mapYelpCookies,
} from "@/lib/playwright"

// Optional cookie mappers keyed by provider. If no mapper exists for a
// provider, cookies are stored as-is (passthrough).
const COOKIE_MAPPERS: Record<string, (c: Record<string, string>) => Record<string, string>> = {
  instagram: mapInstagramCookies,
  linkedin: mapLinkedinCookies,
  x: mapXCookies,
  canvas: mapCanvasCookies,
  zillow: mapZillowCookies,
  streeteasy: mapStreetEasyCookies,
  yelp: mapYelpCookies,
}

export async function POST(request: NextRequest) {
  const csrfError = checkOrigin(request)
  if (csrfError) return csrfError

  const supabase = await createClient()
  const {
    data: { user },
  } = await supabase.auth.getUser()

  if (!user || !isAdmin(user.email)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  let body: { provider?: unknown; label?: unknown; cookies?: unknown; baseUrl?: unknown }
  try {
    body = await request.json()
  } catch {
    return NextResponse.json({ error: "Invalid request body" }, { status: 400 })
  }

  const { provider, label, cookies } = body

  if (!provider || typeof provider !== "string") {
    return NextResponse.json({ error: "Missing provider" }, { status: 400 })
  }

  if (!cookies || typeof cookies !== "object" || Array.isArray(cookies)) {
    return NextResponse.json(
      { error: "cookies must be a JSON object of cookie name → value pairs" },
      { status: 400 }
    )
  }

  const rawCookies = cookies as Record<string, string>
  const accountLabel =
    typeof label === "string" && label.trim()
      ? label.trim()
      : `${provider} Account`

  // Apply provider-specific mapper if one exists; otherwise store as-is
  const mapper = COOKIE_MAPPERS[provider]
  const mapped = mapper ? mapper(rawCookies) : rawCookies

  // Canvas needs the base URL stored alongside cookies
  if (provider === "canvas" && typeof body.baseUrl === "string" && body.baseUrl.trim()) {
    let normalizedUrl = body.baseUrl.trim().replace(/\/+$/, "")
    if (!/^https?:\/\//i.test(normalizedUrl)) {
      normalizedUrl = `https://${normalizedUrl}`
    }
    mapped.base_url = normalizedUrl
  }

  const encrypted = encrypt(JSON.stringify(mapped))
  const credHex = `\\x${Buffer.from(encrypted).toString("hex")}`
  const now = new Date().toISOString()

  const admin = createAdminClient()

  // Upsert by (user_id, provider, label)
  const { data: existing } = await admin
    .from("user_integrations")
    .select("id")
    .eq("user_id", user.id)
    .eq("provider", provider)
    .eq("label", accountLabel)
    .maybeSingle()

  if (existing) {
    const { error } = await admin
      .from("user_integrations")
      .update({ credentials: credHex, status: "active", updated_at: now })
      .eq("id", existing.id)

    if (error) {
      console.error(`[save-cookies] DB update error for ${provider}:`, error)
      return NextResponse.json({ error: "Failed to update credentials" }, { status: 500 })
    }
  } else {
    const { error } = await admin
      .from("user_integrations")
      .insert({
        user_id: user.id,
        provider,
        credentials: credHex,
        status: "active",
        label: accountLabel,
        created_at: now,
        updated_at: now,
      })

    if (error) {
      console.error(`[save-cookies] DB insert error for ${provider}:`, error)
      return NextResponse.json({ error: "Failed to save credentials" }, { status: 500 })
    }
  }

  return NextResponse.json({ ok: true, label: accountLabel })
}
