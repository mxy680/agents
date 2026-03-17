import { createCipheriv, createDecipheriv, randomBytes } from "crypto";

const ALGORITHM = "aes-256-gcm";
const NONCE_LENGTH = 12;
const TAG_LENGTH = 16;

function getKey(): Buffer {
  const hex = process.env.PORTAL_ENCRYPTION_KEY;
  if (!hex || hex.length !== 64) {
    throw new Error(
      "PORTAL_ENCRYPTION_KEY must be 64 hex characters (32 bytes)"
    );
  }
  return Buffer.from(hex, "hex");
}

/**
 * Encrypt plaintext with AES-256-GCM.
 * Wire format: base64(nonce [12] || ciphertext || tag [16])
 * Compatible with Go crypto/aes GCM.
 */
export function encrypt(plaintext: string): string {
  const key = getKey();
  const nonce = randomBytes(NONCE_LENGTH);
  const cipher = createCipheriv(ALGORITHM, key, nonce);

  const encrypted = Buffer.concat([
    cipher.update(plaintext, "utf8"),
    cipher.final(),
  ]);
  const tag = cipher.getAuthTag();

  return Buffer.concat([nonce, encrypted, tag]).toString("base64");
}

/**
 * Decrypt a value produced by encrypt() or Go's gcm.Seal().
 */
export function decrypt(encoded: string): string {
  const key = getKey();
  const data = Buffer.from(encoded, "base64");

  if (data.length < NONCE_LENGTH + TAG_LENGTH) {
    throw new Error("ciphertext too short");
  }

  const nonce = data.subarray(0, NONCE_LENGTH);
  const tag = data.subarray(data.length - TAG_LENGTH);
  const ciphertext = data.subarray(NONCE_LENGTH, data.length - TAG_LENGTH);

  const decipher = createDecipheriv(ALGORITHM, key, nonce);
  decipher.setAuthTag(tag);

  return Buffer.concat([
    decipher.update(ciphertext),
    decipher.final(),
  ]).toString("utf8");
}
