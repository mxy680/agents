export interface ProviderMeta {
  id: string;
  name: string;
  description: string;
  scopes?: string[];
  authType: "oauth" | "cookies";
}

export const providers: ProviderMeta[] = [
  {
    id: "google",
    name: "Google",
    description: "Gmail, Sheets, Calendar, and Drive",
    scopes: [
      "https://mail.google.com/",
      "https://www.googleapis.com/auth/gmail.settings.basic",
      "https://www.googleapis.com/auth/gmail.settings.sharing",
      "https://www.googleapis.com/auth/spreadsheets",
      "https://www.googleapis.com/auth/drive.file",
      "https://www.googleapis.com/auth/calendar",
      "https://www.googleapis.com/auth/drive",
    ],
    authType: "oauth",
  },
  {
    id: "github",
    name: "GitHub",
    description: "Repos, Issues, PRs, Actions, and Gists",
    scopes: ["repo", "gist", "read:org", "workflow"],
    authType: "oauth",
  },
  {
    id: "instagram",
    name: "Instagram",
    description: "Posts, Stories, Reels, Comments, and DMs",
    authType: "cookies",
  },
];

export function getProvider(id: string): ProviderMeta | undefined {
  return providers.find((p) => p.id === id);
}
