import type { ProviderConfig } from "../session-capture";

export const instagramConfig: ProviderConfig = {
  loginUrl: "https://www.instagram.com/accounts/login/",
  domain: "https://www.instagram.com",
  requiredCookies: ["sessionid", "csrftoken", "ds_user_id"],
  optionalCookies: ["mid", "ig_did"],
  displayName: "Instagram",
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
