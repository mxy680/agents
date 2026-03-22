import { chromium, type BrowserContext, type Cookie } from "playwright";
import { join } from "path";
import { homedir } from "os";

export type SessionStatus =
  | "pending"
  | "browser_open"
  | "waiting_login"
  | "capturing"
  | "done"
  | "error";

export interface SessionProgress {
  sessionId: string;
  provider: string;
  status: SessionStatus;
  message: string;
  cookies?: Record<string, string>;
  error?: string;
}

// In-memory session state (single-process, admin-only)
const sessions = new Map<string, SessionProgress>();

export function getSession(sessionId: string): SessionProgress | undefined {
  return sessions.get(sessionId);
}

function updateSession(sessionId: string, update: Partial<SessionProgress>) {
  const current = sessions.get(sessionId);
  if (current) {
    sessions.set(sessionId, { ...current, ...update });
  }
}

export interface ProviderConfig {
  loginUrl: string;
  domain: string;
  cookieNames: string[];
  /** Optional function to check if user is logged in */
  isLoggedIn?: (ctx: BrowserContext) => Promise<boolean>;
  /** Human-readable provider name */
  displayName: string;
}

const USER_DATA_DIR = join(homedir(), ".emdash", "playwright-profiles");

/**
 * Launch a visible browser for the user to log in, then capture cookies.
 * Runs in a spawned async context — caller should not await directly in request handler.
 */
export async function captureSession(
  provider: string,
  config: ProviderConfig
): Promise<string> {
  const sessionId = `${provider}-${Date.now()}`;
  sessions.set(sessionId, {
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
  const profileDir = join(USER_DATA_DIR, provider);

  updateSession(sessionId, {
    status: "browser_open",
    message: `Opening browser for ${config.displayName}...`,
  });

  const context = await chromium.launchPersistentContext(profileDir, {
    headless: false,
    viewport: { width: 1280, height: 900 },
    locale: "en-US",
    timezoneId: "America/New_York",
    args: ["--disable-blink-features=AutomationControlled"],
  });

  try {
    const page = context.pages()[0] || (await context.newPage());

    updateSession(sessionId, {
      status: "waiting_login",
      message: `Navigate to ${config.displayName} and log in. The browser will close automatically when done.`,
    });

    await page.goto(config.loginUrl, { waitUntil: "domcontentloaded" });

    // Poll for the required cookies to appear
    const maxWaitMs = 5 * 60 * 1000; // 5 minutes
    const pollIntervalMs = 2000;
    let elapsed = 0;

    while (elapsed < maxWaitMs) {
      const cookies = await context.cookies(config.domain);
      const cookieMap = new Map(cookies.map((c) => [c.name, c.value]));
      const allPresent = config.cookieNames.every(
        (name) => cookieMap.has(name) && cookieMap.get(name) !== ""
      );

      if (allPresent) {
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

        const result: Record<string, string> = {};
        for (const name of config.cookieNames) {
          result[name] = cookieMap.get(name) || "";
        }

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
  }
}
