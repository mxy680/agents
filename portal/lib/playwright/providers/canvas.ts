import type { ProviderConfig } from "../session-capture";

/**
 * Canvas config factory — needs the institution's Canvas URL.
 * The base URL is read from CANVAS_BASE_URL env var.
 */
export function canvasConfig(baseUrl: string): ProviderConfig {
  return {
    loginUrl: `${baseUrl}/login`,
    domain: baseUrl,
    cookieNames: ["_normandy_session", "_csrf_token"],
    displayName: "Canvas LMS",
  };
}

export function mapCanvasCookies(
  cookies: Record<string, string>
): Record<string, string> {
  return {
    session_cookie: cookies._normandy_session || "",
    csrf_token: cookies._csrf_token || "",
    log_session_id: cookies.log_session_id || "",
  };
}
