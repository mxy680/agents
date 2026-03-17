import { NextResponse } from "next/server";
import { createClient, createServiceClient } from "@/lib/supabase/server";
import { encryptCredentials } from "@/lib/crypto";

export async function GET(request: Request) {
  const { searchParams, origin } = new URL(request.url);
  const code = searchParams.get("code");
  const state = searchParams.get("state");
  const error = searchParams.get("error");

  if (error || !code) {
    return NextResponse.redirect(
      `${origin}/integrations?error=google_auth_failed`
    );
  }

  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user || user.id !== state) {
    return NextResponse.redirect(`${origin}/login`);
  }

  // Exchange code for tokens
  const tokenRes = await fetch("https://oauth2.googleapis.com/token", {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      code,
      client_id: process.env.GOOGLE_CLIENT_ID!,
      client_secret: process.env.GOOGLE_CLIENT_SECRET!,
      redirect_uri: `${origin}/api/integrations/google/callback`,
      grant_type: "authorization_code",
    }),
  });

  if (!tokenRes.ok) {
    return NextResponse.redirect(
      `${origin}/integrations?error=google_token_exchange_failed`
    );
  }

  const tokens = await tokenRes.json();

  // Fetch Google account email to use as account label
  const userinfoRes = await fetch(
    "https://www.googleapis.com/oauth2/v2/userinfo",
    { headers: { Authorization: `Bearer ${tokens.access_token}` } }
  );
  const userinfo = await userinfoRes.json();
  const accountLabel = userinfo.email ?? "";

  // Look up the google integration ID from catalog
  const { data: integration } = await supabase
    .from("integrations")
    .select("id")
    .eq("name", "google")
    .single();

  if (!integration) {
    return NextResponse.redirect(
      `${origin}/integrations?error=google_integration_not_found`
    );
  }

  // Encrypt all credentials as a single JSON blob → bytea
  const creds: Record<string, string> = {
    access_token: tokens.access_token,
  };
  if (tokens.refresh_token) {
    creds.refresh_token = tokens.refresh_token;
  }
  if (tokens.expires_in) {
    creds.expires_at = new Date(
      Date.now() + tokens.expires_in * 1000
    ).toISOString();
  }

  const serviceClient = await createServiceClient();
  const { error: upsertError } = await serviceClient
    .from("user_integrations")
    .upsert(
      {
        user_id: user.id,
        integration_id: integration.id,
        account_label: accountLabel,
        credentials: encryptCredentials(creds),
        status: "connected",
        connected_at: new Date().toISOString(),
        expires_at: creds.expires_at ?? null,
      },
      { onConflict: "user_id,integration_id,account_label" }
    );

  if (upsertError) {
    return NextResponse.redirect(
      `${origin}/integrations?error=${encodeURIComponent(upsertError.message)}`
    );
  }

  return NextResponse.redirect(`${origin}/integrations`);
}
