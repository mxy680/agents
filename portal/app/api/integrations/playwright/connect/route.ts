import { NextRequest, NextResponse } from "next/server";
import { createClient } from "@/lib/supabase/server";
import { createAdminClient } from "@/lib/supabase/admin";
import { isAdmin } from "@/lib/admin";
import {
  captureSession,
  getSession,
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
      const { error: dbError } = await admin.from("user_integrations").upsert(
        {
          user_id: userId,
          provider,
          status: "active",
          credentials: `\\x${Buffer.from(encrypted).toString("hex")}`,
          updated_at: new Date().toISOString(),
        },
        { onConflict: "user_id,provider" }
      );

      if (dbError) {
        console.error(`[playwright/connect] DB error saving ${provider}:`, dbError);
      }

      return;
    }

    if (session.status === "error") return;

    await new Promise((r) => setTimeout(r, interval));
    elapsed += interval;
  }
}
