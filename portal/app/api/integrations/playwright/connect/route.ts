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
  const { provider } = body as { provider: string };

  if (!provider || !PROVIDER_CONFIGS[provider]) {
    return NextResponse.json(
      {
        error: `Unsupported provider: ${provider}. Supported: ${Object.keys(PROVIDER_CONFIGS).join(", ")}`,
      },
      { status: 400 }
    );
  }

  const { config } = PROVIDER_CONFIGS[provider];
  const sessionId = await captureSession(provider, config());

  // Start background task to save cookies when done
  pollAndSave(sessionId, provider, user.id).catch(console.error);

  return NextResponse.json({ sessionId, provider });
}

async function pollAndSave(
  sessionId: string,
  provider: string,
  userId: string
) {
  const maxWait = 6 * 60 * 1000; // 6 minutes (slightly longer than browser timeout)
  const interval = 1000;
  let elapsed = 0;

  while (elapsed < maxWait) {
    const session = getSession(sessionId);
    if (!session) return;

    if (session.status === "done" && session.cookies) {
      const { mapCookies } = PROVIDER_CONFIGS[provider];
      const mapped = mapCookies(session.cookies);

      const encrypted = encrypt(JSON.stringify(mapped));

      const admin = createAdminClient();
      const credHex = `\\x${Buffer.from(encrypted).toString("hex")}`;
      const now = new Date().toISOString();

      // Try update first (existing row for this user+provider)
      const { data: updated, error: updateError } = await admin
        .from("user_integrations")
        .update({
          credentials: credHex,
          status: "active",
          updated_at: now,
        })
        .eq("user_id", userId)
        .eq("provider", provider)
        .select("id");

      if (updateError) {
        console.error(`[playwright/connect] DB update error for ${provider}:`, updateError);
        updateSessionExternal(sessionId, {
          status: "error",
          message: `Failed to save ${provider} credentials.`,
          error: updateError.message,
        });
      } else if (!updated || updated.length === 0) {
        // No existing row — insert
        const { error: insertError } = await admin
          .from("user_integrations")
          .insert({
            user_id: userId,
            provider,
            credentials: credHex,
            status: "active",
            label: provider,
            created_at: now,
            updated_at: now,
          });

        if (insertError) {
          console.error(`[playwright/connect] DB insert error for ${provider}:`, insertError);
          updateSessionExternal(sessionId, {
            status: "error",
            message: `Failed to save ${provider} credentials.`,
            error: insertError.message,
          });
        } else {
          updateSessionExternal(sessionId, {
            status: "saved",
            message: `${provider} credentials saved successfully.`,
          });
        }
      } else {
        updateSessionExternal(sessionId, {
          status: "saved",
          message: `${provider} credentials updated successfully.`,
        });
      }

      return;
    }

    if (session.status === "error") return;

    await new Promise((r) => setTimeout(r, interval));
    elapsed += interval;
  }
}
