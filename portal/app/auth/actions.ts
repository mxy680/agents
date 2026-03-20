"use server"

import { headers } from "next/headers"
import { redirect } from "next/navigation"
import { createClient } from "@/lib/supabase/server"

/** Derive the site origin from the incoming request so localhost and production both redirect correctly. */
async function getSiteUrl(): Promise<string> {
  const h = await headers()
  const origin = h.get("origin") || h.get("referer")
  if (origin) {
    try {
      const url = new URL(origin)
      return url.origin // e.g. "http://localhost:3000" or "https://agents.markshteyn.com"
    } catch {
      // fall through
    }
  }
  return process.env.NEXT_PUBLIC_SITE_URL || "http://localhost:3000"
}

export async function signInWithGoogle() {
  const siteUrl = await getSiteUrl()
  const supabase = await createClient()
  const { data, error } = await supabase.auth.signInWithOAuth({
    provider: "google",
    options: {
      redirectTo: `${siteUrl}/auth/callback`,
    },
  })

  if (error) {
    redirect(`/login?error=${encodeURIComponent(error.message)}`)
  }

  if (data.url) {
    redirect(data.url)
  }
}

export async function signInWithGitHub() {
  const siteUrl = await getSiteUrl()
  const supabase = await createClient()
  const { data, error } = await supabase.auth.signInWithOAuth({
    provider: "github",
    options: {
      redirectTo: `${siteUrl}/auth/callback`,
      scopes: "user:email",
    },
  })

  if (error) {
    redirect(`/login?error=${encodeURIComponent(error.message)}`)
  }

  if (data.url) {
    redirect(data.url)
  }
}

export async function signOut() {
  const supabase = await createClient()
  await supabase.auth.signOut()
  redirect("/login")
}
