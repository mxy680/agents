import { createHmac, timingSafeEqual } from "crypto"

const STATE_TTL_MS = 10 * 60 * 1000 // 10 minutes

function getKey(): string {
  const hex = process.env.ENCRYPTION_MASTER_KEY
  if (!hex || hex.length !== 64) {
    throw new Error("ENCRYPTION_MASTER_KEY must be a 64-char hex string (32 bytes)")
  }
  return hex
}

function sign(payload: string, key: string): string {
  return createHmac("sha256", Buffer.from(key, "hex")).update(payload).digest("hex")
}

/**
 * Builds an HMAC-SHA256-signed OAuth state token.
 * Format: base64url(payload + "." + signature)
 * where payload = "userId:label:timestampMs"
 */
export function createOAuthState(userId: string, label: string): string {
  const key = getKey()
  const payload = `${userId}:${label}:${Date.now()}`
  const signature = sign(payload, key)
  const combined = `${payload}.${signature}`
  return Buffer.from(combined).toString("base64url")
}

/**
 * Verifies an OAuth state token created by createOAuthState.
 * Throws if the signature is invalid or the token is older than 10 minutes.
 */
export function verifyOAuthState(state: string): { userId: string; label: string } {
  const key = getKey()

  let combined: string
  try {
    combined = Buffer.from(state, "base64url").toString("utf8")
  } catch {
    throw new Error("Invalid state encoding")
  }

  const dotIdx = combined.lastIndexOf(".")
  if (dotIdx === -1) {
    throw new Error("Invalid state format: missing signature")
  }

  const payload = combined.slice(0, dotIdx)
  const receivedSig = combined.slice(dotIdx + 1)

  const expectedSig = sign(payload, key)

  // Constant-time comparison to prevent timing attacks
  const receivedBuf = Buffer.from(receivedSig, "hex")
  const expectedBuf = Buffer.from(expectedSig, "hex")
  if (
    receivedBuf.length !== expectedBuf.length ||
    !timingSafeEqual(receivedBuf, expectedBuf)
  ) {
    throw new Error("Invalid state signature")
  }

  // payload = "userId:label:timestampMs"
  // userId may contain colons (UUIDs don't, but be safe by splitting on last two colons)
  const lastColon = payload.lastIndexOf(":")
  if (lastColon === -1) {
    throw new Error("Invalid state payload: missing timestamp")
  }
  const timestampMs = parseInt(payload.slice(lastColon + 1), 10)
  if (isNaN(timestampMs)) {
    throw new Error("Invalid state payload: bad timestamp")
  }
  if (Date.now() - timestampMs > STATE_TTL_MS) {
    throw new Error("State token expired")
  }

  const rest = payload.slice(0, lastColon)
  const firstColon = rest.indexOf(":")
  if (firstColon === -1) {
    throw new Error("Invalid state payload: missing label separator")
  }

  const userId = rest.slice(0, firstColon)
  const label = rest.slice(firstColon + 1)

  return { userId, label }
}
