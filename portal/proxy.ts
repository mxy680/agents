import { NextResponse, type NextRequest } from "next/server"

export async function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl
  const host = request.headers.get("host") ?? ""
  const isProduction = host.includes("agents.markshteyn.com") || host.includes("fly.dev")

  // Production: only /client and /api/client are accessible
  if (isProduction) {
    if (
      pathname.startsWith("/client") ||
      pathname.startsWith("/api/client") ||
      pathname.startsWith("/api/chat/upload") ||
      pathname.startsWith("/api/jobs")
    ) {
      return NextResponse.next({ request })
    }

    // Everything else → /client
    return NextResponse.redirect(new URL("/client", request.url))
  }

  // Local dev: allow everything, redirect / to /integrations
  if (pathname === "/") {
    return NextResponse.redirect(new URL("/integrations", request.url))
  }

  return NextResponse.next({ request })
}

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|api/integrations/zillow/scrape-results).*)",
  ],
}
