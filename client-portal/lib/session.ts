import { createHmac } from "crypto"
import { cookies } from "next/headers"

const SECRET =
  process.env.SESSION_SECRET ||
  process.env.SUPABASE_SERVICE_ROLE_KEY ||
  "fallback-dev-secret"

/**
 * Sign a value with HMAC-SHA256 so it can't be forged.
 * Format: value.signature
 */
export function signSession(code: string): string {
  const sig = createHmac("sha256", SECRET).update(code).digest("hex")
  return `${code}.${sig}`
}

/**
 * Verify and extract the code from a signed session value.
 * Returns null if the signature is invalid.
 */
export function verifySession(signed: string): string | null {
  const lastDot = signed.lastIndexOf(".")
  if (lastDot === -1) return null

  const code = signed.substring(0, lastDot)
  const sig = signed.substring(lastDot + 1)

  const expected = createHmac("sha256", SECRET).update(code).digest("hex")

  // Constant-time comparison to prevent timing attacks
  if (sig.length !== expected.length) return null
  let match = true
  for (let i = 0; i < sig.length; i++) {
    if (sig[i] !== expected[i]) match = false
  }

  return match ? code : null
}

export async function getSessionCode(): Promise<string | null> {
  const cookieStore = await cookies()
  const signed = cookieStore.get("engagent_session")?.value ?? null
  if (!signed) return null
  return verifySession(signed)
}
