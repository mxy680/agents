/**
 * Validates that a request originates from our own site by checking the Origin header.
 * Returns null if valid, or a Response to return if invalid.
 */
export function checkOrigin(request: Request): Response | null {
  const origin = request.headers.get("origin")
  const siteUrl = process.env.NEXT_PUBLIC_SITE_URL ?? "http://localhost:3000"
  if (!origin || origin !== siteUrl) {
    return Response.json({ error: "Forbidden" }, { status: 403 })
  }
  return null
}
