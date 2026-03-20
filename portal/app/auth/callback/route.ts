import { NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url)
  const baseUrl = process.env.NEXT_PUBLIC_SITE_URL ?? new URL(request.url).origin
  const code = searchParams.get("code")
  const errorParam = searchParams.get("error_description") ?? searchParams.get("error")
  const next = searchParams.get("next") ?? "/integrations"

  if (errorParam) {
    return NextResponse.redirect(
      `${baseUrl}/login?error=${encodeURIComponent(errorParam)}`
    )
  }

  if (code) {
    const supabase = await createClient()
    const { error } = await supabase.auth.exchangeCodeForSession(code)
    if (!error) {
      return NextResponse.redirect(`${baseUrl}${next}`)
    }
    return NextResponse.redirect(
      `${baseUrl}/login?error=${encodeURIComponent(error.message)}`
    )
  }

  return NextResponse.redirect(`${baseUrl}/login?error=auth_callback_failed`)
}
