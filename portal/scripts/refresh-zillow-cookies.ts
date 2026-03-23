#!/usr/bin/env npx tsx
/**
 * Automated Zillow cookie refresh script.
 *
 * Launches a real (non-headless) Chromium browser, visits zillow.com,
 * waits for PerimeterX cookies to appear, then saves them to Supabase.
 *
 * Run via cron or LaunchAgent:
 *   cd portal && doppler run -- npx tsx scripts/refresh-zillow-cookies.ts
 *
 * Requirements:
 *   - Playwright chromium installed (npx playwright install chromium)
 *   - Environment variables: NEXT_PUBLIC_SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY, ENCRYPTION_MASTER_KEY
 *   - A display server (macOS: just needs lid open or external display)
 */

import { chromium } from "playwright"
import { createClient } from "@supabase/supabase-js"
import crypto from "crypto"

const ZILLOW_URL = "https://www.zillow.com/homes/Denver,-CO_rb/"
const REQUIRED_COOKIE = "_pxvid"
const MAX_WAIT_MS = 60_000 // 1 minute
const POLL_MS = 2_000

function encrypt(plaintext: string): Buffer {
  const key = Buffer.from(process.env.ENCRYPTION_MASTER_KEY!, "hex")
  const nonce = crypto.randomBytes(12)
  const cipher = crypto.createCipheriv("aes-256-gcm", key, nonce)
  const encrypted = Buffer.concat([cipher.update(plaintext, "utf8"), cipher.final()])
  const tag = cipher.getAuthTag()
  return Buffer.concat([nonce, encrypted, tag])
}

async function main() {
  console.log("[zillow-refresh] Starting cookie refresh...")

  // Launch real browser (not headless — PerimeterX detects headless)
  const browser = await chromium.launch({
    headless: false,
    args: [
      "--disable-blink-features=AutomationControlled",
      "--window-position=-2000,-2000", // move off-screen so it's invisible
      "--window-size=1280,900",
    ],
  })

  const context = await browser.newContext({
    viewport: { width: 1280, height: 900 },
    locale: "en-US",
    timezoneId: "America/New_York",
  })

  try {
    const page = await context.newPage()
    console.log("[zillow-refresh] Navigating to Zillow...")
    await page.goto(ZILLOW_URL, { waitUntil: "domcontentloaded" })

    // Poll for PerimeterX cookies
    let elapsed = 0
    while (elapsed < MAX_WAIT_MS) {
      const cookies = await context.cookies()
      const cookieMap = new Map(cookies.map((c) => [c.name, c.value]))

      if (cookieMap.has(REQUIRED_COOKIE) && cookieMap.get(REQUIRED_COOKIE) !== "") {
        // Verify we're on a real page (not CAPTCHA)
        const url = page.url()
        if (url.startsWith("https://www.zillow.com") && !url.includes("captcha")) {
          // Filter to zillow.com domain
          const zillowCookies = cookies.filter((c) => {
            const domain = c.domain.startsWith(".") ? c.domain.slice(1) : c.domain
            return "www.zillow.com" === domain || "www.zillow.com".endsWith("." + domain)
          })

          const cookiePairs = zillowCookies
            .filter((c) => c.value)
            .map((c) => `${c.name}=${c.value}`)
            .join("; ")

          console.log(`[zillow-refresh] Captured ${zillowCookies.length} cookies`)

          // Save to Supabase
          await saveCookies(cookiePairs)
          console.log("[zillow-refresh] Cookies saved successfully")
          return
        }
      }

      await new Promise((r) => setTimeout(r, POLL_MS))
      elapsed += POLL_MS
    }

    console.error("[zillow-refresh] Timed out waiting for cookies")
    process.exit(1)
  } finally {
    await context.close()
    await browser.close()
  }
}

async function saveCookies(allCookies: string) {
  const supabase = createClient(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.SUPABASE_SERVICE_ROLE_KEY!
  )

  const credentials = JSON.stringify({ all_cookies: allCookies })
  const encrypted = encrypt(credentials)
  const credHex = `\\x${encrypted.toString("hex")}`
  const now = new Date().toISOString()

  // Find the admin's existing zillow integration
  const { data: existing } = await supabase
    .from("user_integrations")
    .select("id, user_id")
    .eq("provider", "zillow")
    .eq("status", "active")
    .limit(1)
    .maybeSingle()

  if (existing) {
    // Update existing
    const { error } = await supabase
      .from("user_integrations")
      .update({ credentials: credHex, updated_at: now })
      .eq("id", existing.id)
    if (error) {
      console.error("[zillow-refresh] DB update error:", error)
      process.exit(1)
    }
  } else {
    // Find admin user (first user with integrations)
    const { data: adminRow } = await supabase
      .from("user_integrations")
      .select("user_id")
      .limit(1)
      .single()

    if (!adminRow) {
      console.error("[zillow-refresh] No admin user found. Connect any integration via portal first.")
      process.exit(1)
    }

    const { error } = await supabase
      .from("user_integrations")
      .insert({
        user_id: adminRow.user_id,
        provider: "zillow",
        label: "Zillow",
        status: "active",
        credentials: credHex,
        created_at: now,
        updated_at: now,
      })
    if (error) {
      console.error("[zillow-refresh] DB insert error:", error)
      process.exit(1)
    }
  }
}

main().catch((err) => {
  console.error("[zillow-refresh] Fatal:", err)
  process.exit(1)
})
