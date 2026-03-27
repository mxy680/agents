import { createServerClient } from "@supabase/ssr"
import { cookies } from "next/headers"

/**
 * Creates a Supabase client for server-side use.
 * Auth checks are bypassed — admin dashboard is local-only.
 * The returned client wraps getUser() to always return a local admin user.
 */
export async function createClient() {
  const cookieStore = await cookies()

  const client = createServerClient(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!,
    {
      cookies: {
        getAll() {
          return cookieStore.getAll()
        },
        setAll(cookiesToSet) {
          try {
            cookiesToSet.forEach(({ name, value, options }) =>
              cookieStore.set(name, value, options)
            )
          } catch {
            // Server component — cookie writes are ignored
          }
        },
      },
    }
  )

  // Override auth.getUser to return a local admin user
  const originalAuth = client.auth
  client.auth = {
    ...originalAuth,
    getUser: async () => ({
      data: {
        user: {
          id: "00000000-0000-0000-0000-000000000001",
          email: "admin@localhost",
          user_metadata: { full_name: "Admin" },
          app_metadata: {},
          aud: "authenticated",
          created_at: new Date().toISOString(),
        } as any,
      },
      error: null,
    }),
    getSession: async () => ({
      data: { session: null },
      error: null,
    }),
  } as any

  return client
}
