import { NextResponse } from "next/server";
import { createClient } from "@/lib/supabase/server";
import { encryptCredentials } from "@/lib/crypto";

interface InstagramCookies {
  sessionId: string;
  csrfToken: string;
  dsUserId: string;
  mid?: string;
  igDid?: string;
}

export async function POST(request: Request) {
  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  let body: InstagramCookies;
  try {
    body = await request.json();
  } catch {
    return NextResponse.json({ error: "Invalid JSON" }, { status: 400 });
  }

  if (!body.sessionId || !body.csrfToken || !body.dsUserId) {
    return NextResponse.json(
      { error: "sessionId, csrfToken, and dsUserId are required" },
      { status: 400 }
    );
  }

  const { data: integration } = await supabase
    .from("integrations")
    .select("id")
    .eq("name", "instagram")
    .single();

  if (!integration) {
    return NextResponse.json(
      { error: "Integration not found" },
      { status: 404 }
    );
  }

  const creds: Record<string, string> = {
    session_id: body.sessionId,
    csrf_token: body.csrfToken,
    ds_user_id: body.dsUserId,
  };
  if (body.mid) creds.mid = body.mid;
  if (body.igDid) creds.ig_did = body.igDid;

  const { error } = await supabase.from("user_integrations").upsert(
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

  if (error) {
    return NextResponse.json({ error: error.message }, { status: 500 });
  }

  return NextResponse.json({ success: true });
}
