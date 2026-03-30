import { NextResponse, type NextRequest } from "next/server"

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl
  const hasSession = request.cookies.has("engagent_session")

  // Root → redirect based on session
  if (pathname === "/") {
    return NextResponse.redirect(new URL(hasSession ? "/chat" : "/auth", request.url))
  }

  // Authenticated user trying to access auth page → redirect to chat
  if (pathname === "/auth" && hasSession) {
    return NextResponse.redirect(new URL("/chat", request.url))
  }

  // Unauthenticated user trying to access chat → redirect to auth
  if (pathname.startsWith("/chat") && !hasSession) {
    return NextResponse.redirect(new URL("/auth", request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: ["/", "/auth", "/chat/:path*"],
}
