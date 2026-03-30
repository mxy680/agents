import { NextResponse, type NextRequest } from "next/server"

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl
  const hasSession = request.cookies.has("engagent_session")

  // Authenticated user trying to access sign-in page → redirect to chat
  if (pathname === "/" && hasSession) {
    return NextResponse.redirect(new URL("/chat", request.url))
  }

  // Unauthenticated user trying to access chat → redirect to sign-in
  if (pathname.startsWith("/chat") && !hasSession) {
    return NextResponse.redirect(new URL("/", request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: ["/", "/chat/:path*"],
}
