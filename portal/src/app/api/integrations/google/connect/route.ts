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

  const label = new URL(request.url).searchParams.get("label") ?? "";
  const provider = getProvider("google")!;
  const params = new URLSearchParams({
    client_id: process.env.GOOGLE_CLIENT_ID!,
    redirect_uri: `${new URL(request.url).origin}/api/integrations/google/callback`,
    response_type: "code",
    scope: provider.scopes!.join(" "),
    access_type: "offline",
    prompt: "consent",
    state: `${user.id}:${label}`,
  });

  return NextResponse.redirect(
    `https://accounts.google.com/o/oauth2/v2/auth?${params.toString()}`
  );
}
