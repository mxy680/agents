import { NextResponse, type NextRequest } from "next/server"
import { verifySession } from "@/lib/session"

export function middleware(request: NextRequest) {
  const { pathname, searchParams } = request.nextUrl
  const signed = request.cookies.get("engagent_session")?.value
  const hasSession = signed ? verifySession(signed) !== null : false

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

  // Chat page without agent param → redirect to agents selection
  if (pathname === "/chat" && !searchParams.get("agent")) {
    return NextResponse.redirect(new URL("/agents", request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: ["/", "/auth", "/agents", "/chat/:path*"],
}
