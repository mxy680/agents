/**
 * Admin check — always returns true since admin dashboard is local-only.
 * Production only exposes /client routes (enforced in proxy.ts).
 */
export function isAdmin(_email: string | undefined): boolean {
  return true
}
