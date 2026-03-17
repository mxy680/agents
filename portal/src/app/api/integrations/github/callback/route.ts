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
      `${origin}/integrations?error=github_auth_failed`
    );
  }

  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user || user.id !== state) {
    return NextResponse.redirect(`${origin}/login`);
  }

  const tokenRes = await fetch(
    "https://github.com/login/oauth/access_token",
    {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
      body: JSON.stringify({
        client_id: process.env.GITHUB_CLIENT_ID!,
        client_secret: process.env.GITHUB_CLIENT_SECRET!,
        code,
        redirect_uri: `${origin}/api/integrations/github/callback`,
      }),
    }
  );

  if (!tokenRes.ok) {
    return NextResponse.redirect(
      `${origin}/integrations?error=github_token_exchange_failed`
    );
  }

  const tokens = await tokenRes.json();

  if (tokens.error) {
    return NextResponse.redirect(
      `${origin}/integrations?error=github_token_exchange_failed`
    );
  }

  const { data: integration } = await supabase
    .from("integrations")
    .select("id")
    .eq("name", "github")
    .single();

  if (!integration) {
    return NextResponse.redirect(
      `${origin}/integrations?error=github_integration_not_found`
    );
  }

  const creds: Record<string, string> = {
    access_token: tokens.access_token,
  };
  if (tokens.refresh_token) {
    creds.refresh_token = tokens.refresh_token;
  }

  const serviceClient = await createServiceClient();
  const { error: upsertError } = await serviceClient
    .from("user_integrations")
    .upsert(
      {
        user_id: user.id,
        integration_id: integration.id,
        account_label: "",
        credentials: encryptCredentials(creds),
        status: "connected",
        connected_at: new Date().toISOString(),
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
