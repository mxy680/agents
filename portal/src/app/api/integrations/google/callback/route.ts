import { NextResponse } from "next/server";
import { createClient } from "@/lib/supabase/server";
import { encrypt } from "@/lib/crypto";

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

  const { error: upsertError } = await supabase.from("integrations").upsert(
    {
      user_id: user.id,
      provider: "google",
      status: "active",
      access_token: encrypt(tokens.access_token),
      refresh_token: tokens.refresh_token
        ? encrypt(tokens.refresh_token)
        : null,
      token_expiry: tokens.expires_in
        ? new Date(Date.now() + tokens.expires_in * 1000).toISOString()
        : null,
    },
    { onConflict: "user_id,provider" }
  );

  if (upsertError) {
    return NextResponse.redirect(
      `${origin}/integrations?error=google_save_failed`
    );
  }

  return NextResponse.redirect(`${origin}/integrations`);
}
