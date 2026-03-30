import { NextResponse, type NextRequest } from "next/server"

// Lightweight session check for Edge runtime — just verify cookie format (has a dot separator)
// Full HMAC verification happens in each API route on the Node.js runtime
function hasValidSessionCookie(request: NextRequest): boolean {
  const signed = request.cookies.get("engagent_session")?.value
  if (!signed) return false
  // Signed cookies have format: CODE.HMAC_HEX (64 char hex after the last dot)
  const lastDot = signed.lastIndexOf(".")
  if (lastDot === -1) return false
  const sig = signed.substring(lastDot + 1)
  return sig.length === 64 && /^[0-9a-f]+$/.test(sig)
}

export function middleware(request: NextRequest) {
  const { pathname, searchParams } = request.nextUrl
  const hasSession = hasValidSessionCookie(request)

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
