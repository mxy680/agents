import { createCipheriv, createDecipheriv, randomBytes } from "crypto"

const ALGORITHM = "aes-256-gcm"
const NONCE_BYTES = 12
const TAG_BYTES = 16

function getKey(): Buffer {
  const hex = process.env.ENCRYPTION_MASTER_KEY
  if (!hex || hex.length !== 64) {
    throw new Error("ENCRYPTION_MASTER_KEY must be a 64-char hex string (32 bytes)")
  }
  return Buffer.from(hex, "hex")
}

/**
 * Encrypts a plaintext string using AES-256-GCM.
 * Wire format: nonce [12 bytes] || ciphertext || auth_tag [16 bytes]
 * Compatible with the Go tokenbridge crypto.go implementation.
 */
export function encrypt(plaintext: string): Buffer {
  const key = getKey()
  const nonce = randomBytes(NONCE_BYTES)
  const cipher = createCipheriv(ALGORITHM, key, nonce)

  const encrypted = Buffer.concat([
    cipher.update(plaintext, "utf8"),
    cipher.final(),
  ])
  const tag = cipher.getAuthTag()

  return Buffer.concat([nonce, encrypted, tag])
}

/**
 * Decrypts a Buffer produced by encrypt().
 * Wire format: nonce [12 bytes] || ciphertext || auth_tag [16 bytes]
 */
export function decrypt(ciphertext: Buffer): string {
  const key = getKey()
  const nonce = ciphertext.subarray(0, NONCE_BYTES)
  const tag = ciphertext.subarray(ciphertext.length - TAG_BYTES)
  const data = ciphertext.subarray(NONCE_BYTES, ciphertext.length - TAG_BYTES)

  const decipher = createDecipheriv(ALGORITHM, key, nonce)
  decipher.setAuthTag(tag)

  return Buffer.concat([decipher.update(data), decipher.final()]).toString("utf8")
}
