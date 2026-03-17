import { NextResponse } from "next/server";
import { createClient } from "@/lib/supabase/server";
import { getProvider } from "@/lib/providers";

export async function GET(request: Request) {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  const provider = getProvider("github")!;
  const authorizeParams = new URLSearchParams({
    client_id: process.env.GITHUB_CLIENT_ID!,
    redirect_uri: `${new URL(request.url).origin}/api/integrations/github/callback`,
    scope: provider.scopes!.join(" "),
    state: user.id,
  });

  // Force GitHub to show the login page (allows account switching).
  // GitHub doesn't support prompt=select_account, so we redirect through
  // /login?return_to=/login/oauth/authorize?... which shows the sign-in UI.
  const loginParams = new URLSearchParams({
    return_to: `/login/oauth/authorize?${authorizeParams.toString()}`,
  });

  return NextResponse.redirect(
    `https://github.com/login?${loginParams.toString()}`
  );
}
