import { chromium, type BrowserContext } from "playwright";
import {
  getSession,
  setSession,
  updateSession,
  type SessionProgress,
  type SessionStatus,
} from "./session-store";

// Re-export types and store functions for convenience
export type { SessionProgress, SessionStatus };
export { getSession, updateSession as updateSessionExternal };

export interface ProviderConfig {
  loginUrl: string;
  domain: string;
  /** Cookies that MUST be present to consider login successful */
  requiredCookies: string[];
  /** Cookies to capture if present, but not required */
  optionalCookies?: string[];
  /** Optional function to check if user is logged in */
  isLoggedIn?: (ctx: BrowserContext) => Promise<boolean>;
  /** Human-readable provider name */
  displayName: string;
}

/**
 * Launch a visible browser for the user to log in, then capture cookies.
 * Uses a fresh (incognito) context so old sessions don't interfere.
 * Runs in a spawned async context — caller should not await directly in request handler.
 */
export async function captureSession(
  provider: string,
  config: ProviderConfig
): Promise<string> {
  const sessionId = `${provider}-${Date.now()}`;
  setSession(sessionId, {
    sessionId,
    provider,
    status: "pending",
    message: `Starting ${config.displayName} session capture...`,
  });

  // Run async — don't block caller
  runCapture(sessionId, provider, config).catch((err) => {
    updateSession(sessionId, {
      status: "error",
      message: `Failed: ${err instanceof Error ? err.message : String(err)}`,
      error: err instanceof Error ? err.message : String(err),
    });
  });

  return sessionId;
}

async function runCapture(
  sessionId: string,
  provider: string,
  config: ProviderConfig
) {
  updateSession(sessionId, {
    status: "browser_open",
    message: `Opening browser for ${config.displayName}...`,
  });

  // Fresh incognito context — no old sessions carry over
  const browser = await chromium.launch({
    headless: false,
    args: ["--disable-blink-features=AutomationControlled"],
  });

  const context = await browser.newContext({
    viewport: { width: 1280, height: 900 },
    locale: "en-US",
    timezoneId: "America/New_York",
  });

  try {
    const page = await context.newPage();

    updateSession(sessionId, {
      status: "waiting_login",
      message: `Log in to ${config.displayName} in the browser window. It will close automatically when done.`,
    });

    await page.goto(config.loginUrl, { waitUntil: "domcontentloaded" });

    // Poll for the required cookies to appear
    const maxWaitMs = 5 * 60 * 1000; // 5 minutes
    const pollIntervalMs = 2000;
    let elapsed = 0;
    const allCookieNames = [
      ...config.requiredCookies,
      ...(config.optionalCookies ?? []),
    ];

    while (elapsed < maxWaitMs) {
      // Get cookies for all URLs the browser has visited (not just the config domain)
      const cookies = await context.cookies();
      const cookieMap = new Map(cookies.map((c) => [c.name, c.value]));

      // Log what we see for debugging
      const found = config.requiredCookies.filter(
        (n) => cookieMap.has(n) && cookieMap.get(n) !== ""
      );
      if (elapsed % 10000 < pollIntervalMs) {
        const allNames = cookies.map((c) => `${c.name}@${c.domain}`);
        console.log(
          `[playwright] ${provider}: ${found.length}/${config.requiredCookies.length} required (${found.join(", ") || "none"}). All cookies: ${allNames.join(", ") || "none"}`
        );
      }

      const allRequiredPresent = config.requiredCookies.every(
        (name) => cookieMap.has(name) && cookieMap.get(name) !== ""
      );

      if (allRequiredPresent) {
        // Optional: verify login via custom check
        if (config.isLoggedIn) {
          const loggedIn = await config.isLoggedIn(context);
          if (!loggedIn) {
            await new Promise((r) => setTimeout(r, pollIntervalMs));
            elapsed += pollIntervalMs;
            continue;
          }
        }

        updateSession(sessionId, {
          status: "capturing",
          message: "Cookies detected, capturing session...",
        });

        // Filter cookies to the provider's domain — the browser may have cookies
        // from SSO redirects (LinkedIn, Google, YouTube, etc.) that we don't want
        const providerHost = new URL(config.domain).hostname;
        const domainCookies = cookies.filter((c) => {
          const cookieDomain = c.domain.startsWith(".")
            ? c.domain.slice(1)
            : c.domain;
          return (
            providerHost === cookieDomain ||
            providerHost.endsWith("." + cookieDomain)
          );
        });

        const result: Record<string, string> = {};
        for (const cookie of domainCookies) {
          if (cookie.value) result[cookie.name] = cookie.value;
        }

        console.log(
          `[playwright] ${provider}: captured ${Object.keys(result).length} cookies: ${Object.keys(result).join(", ")}`
        );

        updateSession(sessionId, {
          status: "done",
          message: `${config.displayName} session captured successfully.`,
          cookies: result,
        });

        return;
      }

      await new Promise((r) => setTimeout(r, pollIntervalMs));
      elapsed += pollIntervalMs;
    }

    updateSession(sessionId, {
      status: "error",
      message: `Timed out waiting for ${config.displayName} login (5 minutes).`,
      error: "timeout",
    });
  } finally {
    await context.close();
    await browser.close();
  }
}
