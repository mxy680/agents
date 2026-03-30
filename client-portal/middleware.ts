import { NextResponse, type NextRequest } from "next/server"

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl
  const hasSession = request.cookies.has("engagent_session")

  // Root → redirect based on session
  if (pathname === "/") {
    return NextResponse.redirect(new URL(hasSession ? "/agents" : "/auth", request.url))
  }

  // Authenticated user trying to access auth page → redirect to agents
  if (pathname === "/auth" && hasSession) {
    return NextResponse.redirect(new URL("/agents", request.url))
  }

  // Unauthenticated user trying to access protected pages → redirect to auth
  if ((pathname.startsWith("/chat") || pathname.startsWith("/agents")) && !hasSession) {
    return NextResponse.redirect(new URL("/auth", request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: ["/", "/auth", "/agents", "/chat/:path*"],
}
