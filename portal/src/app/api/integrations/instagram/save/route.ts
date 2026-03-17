import { NextResponse } from "next/server";
import { createClient } from "@/lib/supabase/server";
import { encrypt } from "@/lib/crypto";

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

  const encryptedMetadata: Record<string, string> = {
    session_id: encrypt(body.sessionId),
    csrf_token: encrypt(body.csrfToken),
    ds_user_id: encrypt(body.dsUserId),
  };

  if (body.mid) {
    encryptedMetadata.mid = encrypt(body.mid);
  }
  if (body.igDid) {
    encryptedMetadata.ig_did = encrypt(body.igDid);
  }

  const { error } = await supabase.from("integrations").upsert(
    {
      user_id: user.id,
      provider: "instagram",
      status: "active",
      metadata: encryptedMetadata,
    },
    { onConflict: "user_id,provider" }
  );

  if (error) {
    return NextResponse.json({ error: error.message }, { status: 500 });
  }

  return NextResponse.json({ success: true });
}
