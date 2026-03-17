import { NextResponse } from "next/server";
import { createClient, createServiceClient } from "@/lib/supabase/server";

export async function PATCH(
  request: Request,
  { params }: { params: Promise<{ provider: string }> }
) {
  const { provider } = await params;

  const supabase = await createClient();
  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 });
  }

  let body: { old_label: string; new_label: string };
  try {
    body = await request.json();
  } catch {
    return NextResponse.json({ error: "Invalid JSON" }, { status: 400 });
  }

  if (body.new_label === undefined) {
    return NextResponse.json({ error: "new_label is required" }, { status: 400 });
  }

  const { data: integration } = await supabase
    .from("integrations")
    .select("id")
    .eq("name", provider)
    .single();

  if (!integration) {
    return NextResponse.json({ error: "Integration not found" }, { status: 404 });
  }

  const serviceClient = await createServiceClient();
  const { error } = await serviceClient
    .from("user_integrations")
    .update({ account_label: body.new_label })
    .eq("user_id", user.id)
    .eq("integration_id", integration.id)
    .eq("account_label", body.old_label);

  if (error) {
    return NextResponse.json({ error: error.message }, { status: 500 });
  }

  return NextResponse.json({ success: true });
}
