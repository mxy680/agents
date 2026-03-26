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
  const host = request.headers.get("host") ?? ""
  const isProduction = host.includes("agents.markshteyn.com") || host.includes("fly.dev")

  // On production, redirect all non-client, non-admin routes to /client
  if (isProduction && !pathname.startsWith("/client") && !pathname.startsWith("/api/client") && !pathname.startsWith("/api/chat/upload")) {
    // Allow admin login flow
    if (PUBLIC_ROUTES.some((r) => pathname.startsWith(r))) {
      // Only allow login for admins, redirect everyone else to /client
      if (pathname === "/login" && user && isAdminEmail(user.email)) {
        return NextResponse.redirect(new URL("/integrations", request.url))
      }
      if (pathname === "/login") {
        return supabaseResponse // Allow login page for admin auth
      }
      return supabaseResponse
    }

    // Admin routes — require admin auth
    if (user && isAdminEmail(user.email)) {
      if (pathname === "/") {
        return NextResponse.redirect(new URL("/integrations", request.url))
      }
      return supabaseResponse // Admin can access everything
    }

    // Everyone else → /client
    return NextResponse.redirect(new URL("/client", request.url))
  }

  // Local dev — normal admin routing
  if (pathname === "/") {
    if (user && isAdminEmail(user.email)) {
      return NextResponse.redirect(new URL("/integrations", request.url))
    }
    return NextResponse.redirect(new URL("/login", request.url))
  }

  if (PUBLIC_ROUTES.some((r) => pathname.startsWith(r))) {
    if (pathname === "/login" && user && isAdminEmail(user.email)) {
      return NextResponse.redirect(new URL("/integrations", request.url))
    }
    return supabaseResponse
  }

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
    "/((?!_next/static|_next/image|favicon.ico|auth/callback|api/integrations/zillow/scrape-results|client|api/client|api/chat/upload).*)",
  ],
}
