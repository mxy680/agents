export const maxDuration = 300 // 5 minutes

import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { isAdmin } from "@/lib/admin"
import { encrypt } from "@/lib/crypto"
import { chromium } from "playwright"

const STREETEASY_URL = "https://streeteasy.com/for-sale/nyc"
const MAX_WAIT_MS = 120_000
const PROVIDER = "streeteasy"

export async function POST(request: NextRequest) {
  const supabase = await createClient()
  const {
    data: { user },
  } = await supabase.auth.getUser()

  if (!user || !isAdmin(user.email)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  try {
    const cookies = await captureCookies()
    await saveCookies(user.id, cookies)
    return NextResponse.json({ ok: true, cookieCount: cookies.split("; ").length })
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    return NextResponse.json({ ok: false, error: message }, { status: 500 })
  }
}

async function captureCookies(): Promise<string> {
  const browser = await chromium.launch({
    headless: false,
    args: ["--disable-blink-features=AutomationControlled"],
  })

  const context = await browser.newContext({
    viewport: { width: 1280, height: 900 },
    locale: "en-US",
    timezoneId: "America/New_York",
  })

  try {
    const page = await context.newPage()
    await page.goto(STREETEASY_URL, { waitUntil: "domcontentloaded" })

    await page.waitForEvent("close", { timeout: MAX_WAIT_MS }).catch(() => {})

    const cookies = await context.cookies()
    const seCookies = cookies.filter((c) => {
      const domain = c.domain.startsWith(".") ? c.domain.slice(1) : c.domain
      return "streeteasy.com" === domain || "streeteasy.com".endsWith("." + domain)
    })

    if (seCookies.length === 0) {
      throw new Error("No StreetEasy cookies captured")
    }

    return seCookies
      .filter((c) => c.value)
      .map((c) => `${c.name}=${c.value}`)
      .join("; ")
  } finally {
    await context.close()
    await browser.close()
  }
}

async function saveCookies(userId: string, allCookies: string) {
  const credentials = JSON.stringify({ all_cookies: allCookies })
  const encrypted = encrypt(credentials)
  const credHex = `\\x${Buffer.from(encrypted).toString("hex")}`
  const now = new Date().toISOString()

  const admin = createAdminClient()

  const { data: existing } = await admin
    .from("user_integrations")
    .select("id")
    .eq("provider", PROVIDER)
    .eq("user_id", userId)
    .eq("status", "active")
    .limit(1)
    .maybeSingle()

  if (existing) {
    await admin
      .from("user_integrations")
      .update({ credentials: credHex, updated_at: now })
      .eq("id", existing.id)
  } else {
    await admin
      .from("user_integrations")
      .insert({
        user_id: userId,
        provider: PROVIDER,
        label: "StreetEasy",
        status: "active",
        credentials: credHex,
        created_at: now,
        updated_at: now,
      })
  }
}
