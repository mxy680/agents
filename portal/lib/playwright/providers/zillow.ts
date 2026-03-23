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
    const url = page.url();

    // Must be on zillow.com
    if (!url.startsWith("https://www.zillow.com")) return false;

    // Check that real page content loaded (not CAPTCHA/block page)
    // The CAPTCHA page title is "Access to this page has been denied"
    const title = await page.title();
    if (title.includes("denied") || title.includes("blocked")) return false;

    // Verify actual listings are present (search results page has property cards)
    const hasContent = await page.evaluate(() => {
      return document.querySelectorAll('article, [data-test="property-card"]').length > 0
        || document.querySelector('#grid-search-results') !== null
        || document.querySelector('[id="search-page-list-container"]') !== null;
    }).catch(() => false);

    return hasContent;
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
