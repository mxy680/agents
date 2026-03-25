import type { ProviderConfig } from "../session-capture";

export const yelpConfig: ProviderConfig = {
  loginUrl: "https://www.yelp.com/login",
  domain: "https://www.yelp.com",
  requiredCookies: ["bse"],
  optionalCookies: ["zss", "csrftok"],
  displayName: "Yelp",
  /**
   * Yelp sets the bse cookie before the login redirect completes.
   * Verify the user has landed on a logged-in page (not /login).
   */
  isLoggedIn: async (ctx) => {
    const pages = ctx.pages();
    const page = pages[0];
    if (!page) return false;

    const url = page.url();
    if (
      url.includes("/login") ||
      url.includes("/signup") ||
      url.includes("/account/") ||
      url.includes("/oauth2/")
    ) {
      return false;
    }
    return true;
  },
};

/** Map captured cookies to integration credential keys.
 * Stores ALL cookies as a semicolon-separated string (like Zillow/StreetEasy)
 * because Yelp's DataDome CAPTCHA protection requires the datadome cookie
 * and other tracking cookies to be present.
 */
export function mapYelpCookies(
  cookies: Record<string, string>
): Record<string, string> {
  const allPairs = Object.entries(cookies)
    .map(([k, v]) => `${k}=${v}`)
    .join("; ");
  return {
    all_cookies: allPairs,
    bse: cookies.bse || "",
    zss: cookies.zss || "",
    csrftok: cookies.csrftok || "",
  };
}
