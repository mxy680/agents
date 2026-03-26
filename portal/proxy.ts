import { createServerClient } from "@supabase/ssr"
import { NextResponse, type NextRequest } from "next/server"

const ADMIN_EMAILS = (process.env.ADMIN_EMAILS ?? "")
  .split(",")
  .map((e) => e.trim())
  .filter(Boolean)

function isAdminEmail(email: string | undefined): boolean {
  return !!email && ADMIN_EMAILS.includes(email)
}

// Routes that don't require admin access
const PUBLIC_ROUTES = ["/login", "/auth/callback", "/auth/sign-out", "/api/jobs/cron"]

export async function proxy(request: NextRequest) {
  let supabaseResponse = NextResponse.next({ request })

  const supabase = createServerClient(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!,
    {
      cookies: {
        getAll() {
          return request.cookies.getAll()
        },
        setAll(cookiesToSet) {
          cookiesToSet.forEach(({ name, value }) =>
            request.cookies.set(name, value)
          )
          supabaseResponse = NextResponse.next({ request })
          cookiesToSet.forEach(({ name, value, options }) =>
            supabaseResponse.cookies.set(name, value, options)
          )
        },
      },
    }
  )

  const {
    data: { user },
  } = await supabase.auth.getUser()

  const { pathname } = request.nextUrl

  // Redirect / based on auth state
  if (pathname === "/") {
    if (user && isAdminEmail(user.email)) {
      return NextResponse.redirect(new URL("/integrations", request.url))
    }
    return NextResponse.redirect(new URL("/login", request.url))
  }

  // Allow public routes
  if (PUBLIC_ROUTES.some((r) => pathname.startsWith(r))) {
    // Redirect authenticated admin users away from login
    if (pathname === "/login" && user && isAdminEmail(user.email)) {
      return NextResponse.redirect(new URL("/integrations", request.url))
    }
    return supabaseResponse
  }

  // All other routes require authenticated admin
  if (!user) {
    return NextResponse.redirect(new URL("/login", request.url))
  }

  if (!isAdminEmail(user.email)) {
    return NextResponse.redirect(new URL("/login?error=not_authorized", request.url))
  }

  return supabaseResponse
}

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|auth/callback|api/integrations/zillow/scrape-results).*)",
  ],
}
