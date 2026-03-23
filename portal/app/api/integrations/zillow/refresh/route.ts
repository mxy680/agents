import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { createAdminClient } from "@/lib/supabase/admin"
import { isAdmin } from "@/lib/admin"
import { encrypt } from "@/lib/crypto"
import { chromium } from "playwright"

const ZILLOW_URL = "https://www.zillow.com/homes/Denver,-CO_rb/"
const MAX_WAIT_MS = 120_000 // 2 minutes for user to solve CAPTCHA
const POLL_MS = 2_000

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
    args: [
      "--disable-blink-features=AutomationControlled",
    ],
  })

  const context = await browser.newContext({
    viewport: { width: 1280, height: 900 },
    locale: "en-US",
    timezoneId: "America/New_York",
  })

  try {
    const page = await context.newPage()
    await page.goto(ZILLOW_URL, { waitUntil: "domcontentloaded" })

    let elapsed = 0
    while (elapsed < MAX_WAIT_MS) {
      const cookies = await context.cookies()
      const cookieMap = new Map(cookies.map((c) => [c.name, c.value]))

      // _px3 is set after CAPTCHA is solved
      const hasPx3 = cookieMap.has("_px3") && cookieMap.get("_px3") !== ""
      // search cookie is set after real page loads
      const hasSearch = cookieMap.has("search") && cookieMap.get("search") !== ""

      if (hasPx3 && hasSearch) {
        // Verify not on CAPTCHA page
        const title = await page.title().catch(() => "")
        if (!title.includes("denied") && !title.includes("blocked")) {
          // Filter to zillow.com domain
          const zillowCookies = cookies.filter((c) => {
            const domain = c.domain.startsWith(".") ? c.domain.slice(1) : c.domain
            return "www.zillow.com" === domain || "www.zillow.com".endsWith("." + domain)
          })

          return zillowCookies
            .filter((c) => c.value)
            .map((c) => `${c.name}=${c.value}`)
            .join("; ")
        }
      }

      await new Promise((r) => setTimeout(r, POLL_MS))
      elapsed += POLL_MS
    }

    throw new Error("Timed out waiting for Zillow cookies (2 minutes). Did you solve the CAPTCHA?")
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
    .eq("provider", "zillow")
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
        provider: "zillow",
        label: "Zillow",
        status: "active",
        credentials: credHex,
        created_at: now,
        updated_at: now,
      })
  }
}
