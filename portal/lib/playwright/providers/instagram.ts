import type { ProviderConfig } from "../session-capture";

export const instagramConfig: ProviderConfig = {
  loginUrl: "https://www.instagram.com/accounts/login/",
  domain: "https://www.instagram.com",
  requiredCookies: ["sessionid", "csrftoken", "ds_user_id"],
  optionalCookies: ["mid", "ig_did", "rur"],
  displayName: "Instagram",
  /**
   * Instagram sets session cookies BEFORE 2FA completes. We must verify
   * that the user has actually passed all challenge screens by checking
   * the page URL — challenge/checkpoint pages have distinct URL patterns.
   */
  isLoggedIn: async (ctx) => {
    const pages = ctx.pages();
    const page = pages[0];
    if (!page) return false;

    const url = page.url();
    // Still on login, challenge, or checkpoint pages — not done yet
    if (
      url.includes("/accounts/login") ||
      url.includes("/challenge/") ||
      url.includes("/checkpoint/") ||
      url.includes("/accounts/onetap") ||
      url.includes("/two_factor")
    ) {
      return false;
    }

    // Successfully landed on the main feed or a profile page
    return true;
  },
};

/** Map captured cookies to integration credential keys */
export function mapInstagramCookies(
  cookies: Record<string, string>
): Record<string, string> {
  return {
    session_id: cookies.sessionid || "",
    csrf_token: cookies.csrftoken || "",
    ds_user_id: cookies.ds_user_id || "",
    mid: cookies.mid || "",
    ig_did: cookies.ig_did || "",
  };
}
