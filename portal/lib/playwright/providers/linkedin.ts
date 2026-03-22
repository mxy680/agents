import type { ProviderConfig } from "../session-capture";

export const linkedinConfig: ProviderConfig = {
  loginUrl: "https://www.linkedin.com/login",
  domain: "https://www.linkedin.com",
  cookieNames: ["li_at", "JSESSIONID"],
  displayName: "LinkedIn",
};

export function mapLinkedinCookies(
  cookies: Record<string, string>
): Record<string, string> {
  return {
    li_at: cookies.li_at || "",
    jsessionid: cookies.JSESSIONID || "",
  };
}
