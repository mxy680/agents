import { NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"

/**
 * GET /api/integrations
 *
 * Returns the current user's active integrations. Used by the
 * ExtensionConnectDialog to poll for newly-connected providers.
 */
export async function GET() {
  const supabase = await createClient()
  const {
    data: { user },
  } = await supabase.auth.getUser()

  if (!user) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  const { data: integrations, error } = await supabase
    .from("user_integrations")
    .select("id, provider, label, status")
    .eq("user_id", user.id)

  if (error) {
    return NextResponse.json({ error: error.message }, { status: 500 })
  }

  return NextResponse.json({ integrations: integrations ?? [] })
}
