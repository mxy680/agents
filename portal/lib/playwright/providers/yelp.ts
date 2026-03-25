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

/** Map captured cookies to integration credential keys */
export function mapYelpCookies(
  cookies: Record<string, string>
): Record<string, string> {
  return {
    bse: cookies.bse || "",
    zss: cookies.zss || "",
    csrftok: cookies.csrftok || "",
  };
}
