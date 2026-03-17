import { NextResponse } from "next/server";
import { createClient, createServiceClient } from "@/lib/supabase/server";

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
    .eq("name", "google")
    .single();

  if (!integration) {
    return NextResponse.json({ error: "Integration not found" }, { status: 404 });
  }

  const serviceClient = await createServiceClient();
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
