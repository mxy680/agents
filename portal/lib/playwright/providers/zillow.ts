import type { ProviderConfig } from "../session-capture";

/**
 * Zillow session capture config.
 *
 * Zillow doesn't require login — it uses PerimeterX (HUMAN Security) bot detection.
 * The browser just needs to visit zillow.com and pass the CAPTCHA challenge.
 * Once PerimeterX sets its cookies (_px3, _pxvid, etc.), subsequent API calls
 * work from the Go CLI by sending those cookies.
 *
 * We use isLoggedIn to verify the user has loaded a real Zillow page (not blocked).
 */
export const zillowConfig: ProviderConfig = {
  loginUrl: "https://www.zillow.com/homes/Denver,-CO_rb/",
  domain: "https://www.zillow.com",
  requiredCookies: ["_pxvid"],
  optionalCookies: [
    "_px3",
    "_px2",
    "_pxde",
    "pxcts",
    "JSESSIONID",
    "AWSALB",
    "AWSALBCORS",
    "zguid",
    "zgsession",
    "search",
  ],
  displayName: "Zillow",
  isLoggedIn: async (ctx) => {
    const pages = ctx.pages();
    const page = pages[pages.length - 1];
    if (!page) return false;
    const url = page.url();
    // Verify we're on a real Zillow page, not a CAPTCHA/block page
    return (
      url.startsWith("https://www.zillow.com") &&
      !url.includes("captcha") &&
      !url.includes("blocked")
    );
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
