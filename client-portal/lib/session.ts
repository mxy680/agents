import { createHmac, timingSafeEqual } from "crypto"
import { cookies } from "next/headers"

const SECRET = process.env.SESSION_SECRET || process.env.SUPABASE_SERVICE_ROLE_KEY
if (!SECRET) {
  console.error("WARNING: SESSION_SECRET not set — using SUPABASE_SERVICE_ROLE_KEY as fallback")
}
const SIGNING_KEY = SECRET || ""

/**
 * Sign a value with HMAC-SHA256 so it can't be forged.
 * Format: value.signature
 */
export function signSession(code: string): string {
  const sig = createHmac("sha256", SIGNING_KEY).update(code).digest("hex")
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

  const expected = createHmac("sha256", SIGNING_KEY).update(code).digest("hex")

  try {
    if (!timingSafeEqual(Buffer.from(sig, "hex"), Buffer.from(expected, "hex"))) return null
  } catch {
    return null
  }

  return code
}

export async function getSessionCode(): Promise<string | null> {
  const cookieStore = await cookies()
  const signed = cookieStore.get("engagent_session")?.value ?? null
  if (!signed) return null
  return verifySession(signed)
}
