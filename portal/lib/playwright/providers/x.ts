import type { ProviderConfig } from "../session-capture";

export const xConfig: ProviderConfig = {
  loginUrl: "https://x.com/i/flow/login",
  domain: "https://x.com",
  requiredCookies: ["auth_token", "ct0"],
  displayName: "X (Twitter)",
};

export function mapXCookies(
  cookies: Record<string, string>
): Record<string, string> {
  return {
    auth_token: cookies.auth_token || "",
    csrf_token: cookies.ct0 || "",
  };
}
