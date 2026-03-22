import { NextRequest, NextResponse } from "next/server";
import { createClient } from "@/lib/supabase/server";
import { createAdminClient } from "@/lib/supabase/admin";
import { isAdmin } from "@/lib/admin";
import {
  captureSession,
  getSession,
  updateSessionExternal,
  instagramConfig,
  linkedinConfig,
  xConfig,
  canvasConfig,
  mapInstagramCookies,
  mapLinkedinCookies,
  mapXCookies,
  mapCanvasCookies,
  type ProviderConfig,
} from "@/lib/playwright";
import { encrypt } from "@/lib/crypto";

type ProviderEntry = {
  config: () => ProviderConfig;
  mapCookies: (c: Record<string, string>) => Record<string, string>;
};

const PROVIDER_CONFIGS: Record<string, ProviderEntry> = {
  instagram: { config: () => instagramConfig, mapCookies: mapInstagramCookies },
  linkedin: { config: () => linkedinConfig, mapCookies: mapLinkedinCookies },
  x: { config: () => xConfig, mapCookies: mapXCookies },
  canvas: {
    config: () =>
      canvasConfig(
        process.env.CANVAS_BASE_URL || "https://canvas.instructure.com"
      ),
    mapCookies: mapCanvasCookies,
  },
};

export async function POST(request: NextRequest) {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user || !isAdmin(user.email)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  const body = await request.json();
  const { provider, label, baseUrl } = body as {
    provider: string;
    label?: string;
    baseUrl?: string;
  };

  if (!provider || !PROVIDER_CONFIGS[provider]) {
    return NextResponse.json(
      {
        error: `Unsupported provider: ${provider}. Supported: ${Object.keys(PROVIDER_CONFIGS).join(", ")}`,
      },
      { status: 400 }
    );
  }

  const accountLabel = label?.trim() || `${provider} Account`;

  // Canvas requires a base URL
  if (provider === "canvas" && !baseUrl?.trim()) {
    return NextResponse.json(
      { error: "Canvas base URL is required (e.g. https://canvas.university.edu)" },
      { status: 400 }
    );
  }

  const entry = PROVIDER_CONFIGS[provider];
  const providerConfig =
    provider === "canvas"
      ? canvasConfig(baseUrl!.trim().replace(/\/+$/, ""))
      : entry.config();
  const sessionId = await captureSession(provider, providerConfig);

  // Extra metadata to store alongside cookies
  const extraCreds: Record<string, string> = {};
  if (provider === "canvas" && baseUrl) {
    extraCreds.base_url = baseUrl.trim().replace(/\/+$/, "");
  }

  // Start background task to save cookies when done
  pollAndSave(sessionId, provider, user.id, accountLabel, extraCreds).catch(console.error);

  return NextResponse.json({ sessionId, provider });
}

async function pollAndSave(
  sessionId: string,
  provider: string,
  userId: string,
  label: string,
  extraCreds: Record<string, string> = {}
) {
  const maxWait = 6 * 60 * 1000;
  const interval = 1000;
  let elapsed = 0;

  while (elapsed < maxWait) {
    const session = getSession(sessionId);
    if (!session) return;

    if (session.status === "done" && session.cookies) {
      const { mapCookies } = PROVIDER_CONFIGS[provider];
      const mapped = { ...mapCookies(session.cookies), ...extraCreds };

      const encrypted = encrypt(JSON.stringify(mapped));

      const admin = createAdminClient();
      const credHex = `\\x${Buffer.from(encrypted).toString("hex")}`;
      const now = new Date().toISOString();

      // Upsert by (user_id, provider, label) — allows multiple accounts with different labels
      const { data: existing } = await admin
        .from("user_integrations")
        .select("id")
        .eq("user_id", userId)
        .eq("provider", provider)
        .eq("label", label)
        .maybeSingle();

      if (existing) {
        // Update existing account with this label
        const { error } = await admin
          .from("user_integrations")
          .update({ credentials: credHex, status: "active", updated_at: now })
          .eq("id", existing.id);

        if (error) {
          console.error(`[playwright/connect] DB update error for ${provider}:`, error);
          updateSessionExternal(sessionId, {
            status: "error",
            message: `Failed to save ${provider} credentials.`,
            error: error.message,
          });
        } else {
          updateSessionExternal(sessionId, {
            status: "saved",
            message: `${label} credentials updated.`,
          });
        }
      } else {
        // Insert new account
        const { error } = await admin
          .from("user_integrations")
          .insert({
            user_id: userId,
            provider,
            credentials: credHex,
            status: "active",
            label,
            created_at: now,
            updated_at: now,
          });

        if (error) {
          console.error(`[playwright/connect] DB insert error for ${provider}:`, error);
          updateSessionExternal(sessionId, {
            status: "error",
            message: `Failed to save ${provider} credentials.`,
            error: error.message,
          });
        } else {
          updateSessionExternal(sessionId, {
            status: "saved",
            message: `${label} connected successfully.`,
          });
        }
      }

      return;
    }

    if (session.status === "error") return;

    await new Promise((r) => setTimeout(r, interval));
    elapsed += interval;
  }
}
