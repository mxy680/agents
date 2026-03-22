import type { ProviderConfig } from "../session-capture";

/**
 * Canvas config factory — needs the institution's Canvas URL.
 *
 * Canvas session cookie names vary by instance (_normandy_session, canvas_session, etc.)
 * and SSO flows redirect across domains. Instead of requiring a specific session cookie,
 * we only require _csrf_token and use isLoggedIn to verify the user reached the dashboard.
 * All cookies are captured so whatever session cookie exists gets stored.
 */
export function canvasConfig(baseUrl: string): ProviderConfig {
  return {
    loginUrl: `${baseUrl}/login`,
    domain: baseUrl,
    requiredCookies: ["_csrf_token"],
    optionalCookies: ["_normandy_session", "canvas_session", "log_session_id"],
    displayName: "Canvas LMS",
    isLoggedIn: async (ctx) => {
      const pages = ctx.pages();
      const page = pages[pages.length - 1];
      if (!page) return false;
      const url = page.url();
      // After login, Canvas redirects away from /login to /, /dashboard, /courses, etc.
      const onLoginPage =
        url.includes("/login") ||
        url.includes("/saml") ||
        url.includes("/cas/") ||
        url.includes("/adfs/") ||
        url.includes("/idp/");
      return !onLoginPage && url.startsWith(baseUrl);
    },
  };
}

export function mapCanvasCookies(
  cookies: Record<string, string>
): Record<string, string> {
  // Store all captured cookies — the session cookie name varies by instance
  const mapped: Record<string, string> = {};
  if (cookies._csrf_token) mapped.csrf_token = cookies._csrf_token;
  if (cookies._normandy_session) mapped.session_cookie = cookies._normandy_session;
  if (cookies.canvas_session) mapped.session_cookie = cookies.canvas_session;
  if (cookies.log_session_id) mapped.log_session_id = cookies.log_session_id;

  // Store ALL cookies as a raw semicolon-separated string for the Go CLI
  // This ensures we send whatever session cookie Canvas uses, regardless of name
  const allPairs = Object.entries(cookies)
    .map(([k, v]) => `${k}=${v}`)
    .join("; ");
  if (allPairs) mapped.all_cookies = allPairs;

  return mapped;
}
