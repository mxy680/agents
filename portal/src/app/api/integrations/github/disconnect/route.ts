import { NextResponse } from "next/server";
import { createClient, createServiceClient } from "@/lib/supabase/server";
import { decryptCredentials } from "@/lib/crypto";

export async function POST(request: Request) {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  let body: { account_label: string } = { account_label: "" };
  try {
    body = await request.json();
  } catch {
    // use default
  }

  const { data: integration } = await supabase
    .from("integrations")
    .select("id")
    .eq("name", "github")
    .single();

  if (!integration) {
    return NextResponse.json({ error: "Integration not found" }, { status: 404 });
  }

  // Fetch stored credentials so we can revoke the token on GitHub
  const serviceClient = await createServiceClient();
  const { data: userInteg } = await serviceClient
    .from("user_integrations")
    .select("credentials")
    .eq("user_id", user.id)
    .eq("integration_id", integration.id)
    .eq("account_label", body.account_label)
    .single();

  // Revoke the GitHub token so the next connect shows the authorization UI
  if (userInteg?.credentials) {
    try {
      const creds = decryptCredentials(userInteg.credentials as string);
      if (creds.access_token) {
        await fetch(
          `https://api.github.com/applications/${process.env.GITHUB_CLIENT_ID}/token`,
          {
            method: "DELETE",
            headers: {
              "Content-Type": "application/json",
              Authorization: `Basic ${Buffer.from(
                `${process.env.GITHUB_CLIENT_ID}:${process.env.GITHUB_CLIENT_SECRET}`
              ).toString("base64")}`,
              "X-GitHub-Api-Version": "2022-11-28",
            },
            body: JSON.stringify({ access_token: creds.access_token }),
          }
        );
      }
    } catch {
      // Revocation is best-effort; proceed with DB deletion regardless
    }
  }

  const { error } = await serviceClient
    .from("user_integrations")
    .delete()
    .eq("user_id", user.id)
    .eq("integration_id", integration.id)
    .eq("account_label", body.account_label);

  if (error) {
    return NextResponse.json({ error: error.message }, { status: 500 });
  }

  return NextResponse.json({ success: true });
}
