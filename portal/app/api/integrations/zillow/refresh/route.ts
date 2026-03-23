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

    // Inject a "Done" button into the page. The browser stays open
    // until the user clicks it (after solving any CAPTCHA).
    await page.evaluate(() => {
      const btn = document.createElement("button")
      btn.textContent = "✓ Done — Capture Cookies"
      btn.id = "__zillow_done__"
      Object.assign(btn.style, {
        position: "fixed", top: "10px", right: "10px", zIndex: "999999",
        padding: "12px 24px", fontSize: "16px", fontWeight: "bold",
        background: "#22c55e", color: "white", border: "none",
        borderRadius: "8px", cursor: "pointer", boxShadow: "0 4px 12px rgba(0,0,0,0.3)",
      })
      document.body.appendChild(btn)
    })

    // Set up click handler, then wait for the user to click it
    await page.evaluate(() => {
      const btn = document.getElementById("__zillow_done__")
      if (btn) {
        btn.addEventListener("click", () => {
          ;(window as unknown as Record<string, boolean>).__zillow_captured__ = true
          btn.textContent = "Capturing..."
          btn.style.background = "#666"
        })
      }
    })

    try {
      await page.waitForFunction(
        () => (window as unknown as Record<string, boolean>).__zillow_captured__,
        { timeout: MAX_WAIT_MS }
      )
    } catch {
      throw new Error("Timed out (2 minutes). Click the green 'Done' button after solving the CAPTCHA.")
    }

    // Give cookies a moment to settle
    await new Promise((r) => setTimeout(r, 1000))

    // Capture all zillow.com cookies
    const cookies = await context.cookies()
    const zillowCookies = cookies.filter((c) => {
      const domain = c.domain.startsWith(".") ? c.domain.slice(1) : c.domain
      return "www.zillow.com" === domain || "www.zillow.com".endsWith("." + domain)
    })

    if (zillowCookies.length === 0) {
      throw new Error("No Zillow cookies captured")
    }

    return zillowCookies
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
