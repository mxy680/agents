import type { ProviderConfig } from "../session-capture";

/**
 * StreetEasy session capture config.
 *
 * StreetEasy uses PerimeterX (HUMAN Security) bot detection, the same system
 * as Zillow. On first visit, PerimeterX may show a "press and hold" CAPTCHA
 * challenge. The user must solve it in the browser window. Once solved,
 * PerimeterX sets _px3 (the session cookie that proves the challenge was
 * passed).
 *
 * We require _px3 AND SE_VISITOR_ID, and use isLoggedIn to verify real
 * listings loaded (not the CAPTCHA page). This ensures cookies are only
 * captured after the challenge is fully solved.
 */
export const streeteasyConfig: ProviderConfig = {
  loginUrl: "https://streeteasy.com/for-sale/nyc",
  domain: "https://streeteasy.com",
  requiredCookies: ["_px3", "SE_VISITOR_ID"],
  optionalCookies: [
    "_pxvid",
    "pxcts",
    "AWSALB",
    "AWSALBCORS",
    "_se_t",
  ],
  displayName: "StreetEasy",
  isLoggedIn: async (ctx) => {
    const pages = ctx.pages();
    const page = pages[pages.length - 1];
    if (!page) return false;

    // The CAPTCHA page title is "Access to this page has been denied"
    // If _px3 is required and present, PerimeterX challenge was passed.
    // Just verify we're not still on the block page.
    const title = await page.title().catch(() => "");
    return !title.includes("denied") && !title.includes("blocked");
  },
};

export function mapStreetEasyCookies(
  cookies: Record<string, string>
): Record<string, string> {
  // Store ALL cookies as a semicolon-separated string for the Go CLI.
  // PerimeterX uses multiple cookies together — we need to send them all.
  const allPairs = Object.entries(cookies)
    .map(([k, v]) => `${k}=${v}`)
    .join("; ");

  return {
    all_cookies: allPairs,
  };
}
