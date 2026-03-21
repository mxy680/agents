import { createHmac, timingSafeEqual } from "crypto"

const EXTENSION_TOKEN_TTL_MS = 30 * 24 * 60 * 60 * 1000 // 30 days

function getSigningKey(): string {
  const key = process.env.TOKEN_SIGNING_KEY
  if (!key || key.length !== 64) {
    throw new Error("TOKEN_SIGNING_KEY must be a 64-char hex string (32 bytes)")
  }
  return key
}

/**
 * Creates a long-lived HMAC-SHA256-signed token for the Chrome extension.
 * Format: "ext:<userId>:<timestampMs>.<signature>"
 *
 * Unlike OAuth state tokens this is NOT base64url-encoded — it is sent as a
 * Bearer token in the Authorization header, so plain ASCII is fine.
 */
export function createExtensionToken(userId: string): string {
  const key = getSigningKey()
  const timestamp = Date.now()
  const payload = `ext:${userId}:${timestamp}`
  const signature = createHmac("sha256", Buffer.from(key, "hex"))
    .update(payload)
    .digest("hex")
  return `${payload}.${signature}`
}

/**
 * Verifies an extension token produced by createExtensionToken.
 * Throws if the signature is invalid or the token is older than 30 days.
 */
export function verifyExtensionToken(token: string): { userId: string } {
  const key = getSigningKey()

  const dotIndex = token.lastIndexOf(".")
  if (dotIndex === -1) throw new Error("Invalid token format")

  const payload = token.slice(0, dotIndex)
  const receivedSig = token.slice(dotIndex + 1)

  const expectedSig = createHmac("sha256", Buffer.from(key, "hex"))
    .update(payload)
    .digest("hex")

  // Timing-safe comparison to prevent timing attacks
  const receivedBuf = Buffer.from(receivedSig, "hex")
  const expectedBuf = Buffer.from(expectedSig, "hex")
  if (
    receivedBuf.length !== expectedBuf.length ||
    !timingSafeEqual(receivedBuf, expectedBuf)
  ) {
    throw new Error("Invalid token signature")
  }

  // payload = "ext:<userId>:<timestampMs>"
  const parts = payload.split(":")
  if (parts.length !== 3 || parts[0] !== "ext") {
    throw new Error("Invalid token format")
  }

  const userId = parts[1]
  const timestamp = parseInt(parts[2], 10)
  if (isNaN(timestamp)) throw new Error("Invalid token timestamp")

  if (Date.now() - timestamp > EXTENSION_TOKEN_TTL_MS) {
    throw new Error("Extension token expired")
  }

  return { userId }
}
