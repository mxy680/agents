import type { ProviderConfig } from "../session-capture";

/**
 * Zillow session capture config.
 *
 * Zillow uses PerimeterX (HUMAN Security) bot detection. On first visit,
 * PerimeterX may show a "press and hold" CAPTCHA challenge. The user must
 * solve it in the browser window. Once solved, PerimeterX sets _px3 (the
 * session cookie that proves the challenge was passed).
 *
 * We require _px3 AND use isLoggedIn to verify real property listings
 * loaded (not the CAPTCHA page). This ensures cookies are only captured
 * after the challenge is fully solved.
 */
export const zillowConfig: ProviderConfig = {
  loginUrl: "https://www.zillow.com/homes/Denver,-CO_rb/",
  domain: "https://www.zillow.com",
  requiredCookies: ["_px3", "search"],
  optionalCookies: [
    "_pxvid",
    "_px2",
    "_pxde",
    "pxcts",
    "JSESSIONID",
    "AWSALB",
    "AWSALBCORS",
    "zguid",
    "zgsession",
  ],
  displayName: "Zillow",
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

export function mapZillowCookies(
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
