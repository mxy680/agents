import { chromium, Browser, BrowserContext, Page } from "playwright-core"
import { applyStealthScripts } from "./stealth"
import type { SessionStatus, ClientMessage } from "./types"

const SCREENSHOT_INTERVAL_MS = 80 // ~12 FPS for smoother experience
const COOKIE_CHECK_INTERVAL_MS = 2000
const SESSION_TIMEOUT_MS = 10 * 60 * 1000 // 10 minutes

const INSTAGRAM_COOKIES = [
  "sessionid",
  "csrftoken",
  "ds_user_id",
  "mid",
  "ig_did",
]

export class BrowserSession {
  id: string
  userId: string
  label: string

  private browser: Browser | null = null
  private page: Page | null = null
  private context: BrowserContext | null = null
  private screenshotInterval: NodeJS.Timeout | null = null
  private cookieCheckInterval: NodeJS.Timeout | null = null
  private timeoutTimer: NodeJS.Timeout | null = null
  private onFrame: ((data: string) => void) | null = null
  private onStatus: ((status: SessionStatus) => void) | null = null
  private onCookies: ((cookies: Record<string, string>) => void) | null = null
  destroyed = false

  constructor(id: string, userId: string, label: string) {
    this.id = id
    this.userId = userId
    this.label = label
  }

  setHandlers(handlers: {
    onFrame: (data: string) => void
    onStatus: (status: SessionStatus) => void
    onCookies: (cookies: Record<string, string>) => void
  }): void {
    this.onFrame = handlers.onFrame
    this.onStatus = handlers.onStatus
    this.onCookies = handlers.onCookies
  }

  async start(): Promise<void> {
    this.onStatus?.("loading")

    this.browser = await chromium.launch({
      headless: true,
      args: [
        "--disable-blink-features=AutomationControlled",
        "--no-sandbox",
        "--disable-dev-shm-usage",
      ],
    })

    this.context = await this.browser.newContext({
      viewport: { width: 1280, height: 720 },
      userAgent:
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
      locale: "en-US",
    })

    this.page = await this.context.newPage()
    await applyStealthScripts(this.page)

    // Inject visible cursor on every page load/navigation
    await this.context.addInitScript(() => {
      function injectCursor() {
        if (document.getElementById("__remote_cursor__")) return
        const cursor = document.createElement("div")
        cursor.id = "__remote_cursor__"
        Object.assign(cursor.style, {
          position: "fixed",
          top: "0px",
          left: "0px",
          width: "20px",
          height: "20px",
          borderRadius: "50%",
          border: "2px solid rgba(255, 80, 80, 0.9)",
          backgroundColor: "rgba(255, 80, 80, 0.3)",
          pointerEvents: "none",
          zIndex: "2147483647",
          transform: "translate(-50%, -50%)",
          transition: "left 0.05s linear, top 0.05s linear",
        })
        document.body.appendChild(cursor)
      }
      // Inject immediately if body exists, otherwise wait for DOMContentLoaded
      if (document.body) injectCursor()
      else document.addEventListener("DOMContentLoaded", injectCursor)
    })

    await this.page.goto("https://www.instagram.com/accounts/login/", {
      waitUntil: "domcontentloaded",
    })

    this.onStatus?.("ready")

    // Start screenshot loop
    this.screenshotInterval = setInterval(() => {
      this.captureFrame().catch(() => {})
    }, SCREENSHOT_INTERVAL_MS)

    // Start cookie check loop
    this.cookieCheckInterval = setInterval(() => {
      this.checkCookies().catch(() => {})
    }, COOKIE_CHECK_INTERVAL_MS)

    // Session timeout — reset on each input
    this.resetTimeout()
  }

  private resetTimeout(): void {
    if (this.timeoutTimer) clearTimeout(this.timeoutTimer)
    this.timeoutTimer = setTimeout(() => {
      if (!this.destroyed) {
        this.onStatus?.("timeout")
        this.destroy()
      }
    }, SESSION_TIMEOUT_MS)
  }

  async handleInput(msg: ClientMessage): Promise<void> {
    if (this.destroyed || !this.page) return

    // Reset timeout on any user interaction
    this.resetTimeout()

    // Clamp coordinates to viewport bounds
    const clampX = (x: number) => Math.max(0, Math.min(x, 1280))
    const clampY = (y: number) => Math.max(0, Math.min(y, 720))

    // Move visible cursor for any mouse event
    if ("x" in msg && "y" in msg) {
      const cx = clampX(msg.x)
      const cy = clampY(msg.y)
      this.page.evaluate(({ x, y }) => {
        const el = document.getElementById("__remote_cursor__")
        if (el) { el.style.left = x + "px"; el.style.top = y + "px" }
      }, { x: cx, y: cy }).catch(() => {})
    }

    switch (msg.type) {
      case "click":
        await this.page.mouse.click(clampX(msg.x), clampY(msg.y))
        break
      case "mousemove":
        await this.page.mouse.move(clampX(msg.x), clampY(msg.y))
        break
      case "mousedown":
        await this.page.mouse.move(clampX(msg.x), clampY(msg.y))
        await this.page.mouse.down()
        break
      case "mouseup":
        await this.page.mouse.move(clampX(msg.x), clampY(msg.y))
        await this.page.mouse.up()
        break
      case "keydown":
        await this.page.keyboard.down(msg.key)
        break
      case "keyup":
        await this.page.keyboard.up(msg.key)
        break
      case "keypress":
        await this.page.keyboard.type(msg.text)
        break
      case "scroll":
        await this.page.mouse.wheel(msg.deltaX, msg.deltaY)
        break
    }
  }

  async destroy(): Promise<void> {
    if (this.destroyed) return
    this.destroyed = true

    if (this.screenshotInterval) clearInterval(this.screenshotInterval)
    if (this.cookieCheckInterval) clearInterval(this.cookieCheckInterval)
    if (this.timeoutTimer) clearTimeout(this.timeoutTimer)

    try {
      await this.browser?.close()
    } catch {
      // Ignore close errors
    }

    this.browser = null
    this.context = null
    this.page = null
  }

  private async captureFrame(): Promise<void> {
    if (this.destroyed || !this.page) return
    try {
      const screenshot = await this.page.screenshot({ type: "jpeg", quality: 70 })
      const base64 = screenshot.toString("base64")
      this.onFrame?.(base64)
    } catch {
      // Ignore screenshot errors (page may be navigating)
    }
  }

  private async checkCookies(): Promise<void> {
    if (this.destroyed || !this.context) return
    try {
      const cookies = await this.context.cookies("https://www.instagram.com")
      const sessionCookie = cookies.find(
        (c) => c.name === "sessionid" && c.value !== ""
      )
      if (!sessionCookie) return

      // Login detected — extract all relevant cookies
      this.onStatus?.("login_detected")
      if (this.cookieCheckInterval) clearInterval(this.cookieCheckInterval)
      this.cookieCheckInterval = null

      this.onStatus?.("extracting")
      // Map Instagram cookie names to the keys the token bridge expects
      const COOKIE_TO_CRED: Record<string, string> = {
        sessionid: "session_id",
        csrftoken: "csrf_token",
        ds_user_id: "ds_user_id",
        mid: "mid",
        ig_did: "ig_did",
      }
      const cookieMap: Record<string, string> = {}
      for (const name of INSTAGRAM_COOKIES) {
        const found = cookies.find((c) => c.name === name)
        if (found) cookieMap[COOKIE_TO_CRED[name] ?? name] = found.value
      }

      this.onCookies?.(cookieMap)
    } catch {
      // Ignore errors during cookie check
    }
  }
}
