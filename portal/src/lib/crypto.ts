import { createCipheriv, createDecipheriv, randomBytes } from "crypto";

const ALGORITHM = "aes-256-gcm";
const NONCE_LENGTH = 12;
const TAG_LENGTH = 16;

function getKey(): Buffer {
  const hex = process.env.ENCRYPTION_MASTER_KEY;
  if (!hex || hex.length !== 64) {
    throw new Error(
      "ENCRYPTION_MASTER_KEY must be 64 hex characters (32 bytes)"
    );
  }
  return Buffer.from(hex, "hex");
}

/**
 * Encrypt plaintext with AES-256-GCM.
 * Returns raw bytes: nonce [12] || ciphertext || tag [16]
 * Compatible with Go crypto/aes GCM.
 */
export function encryptToBytes(plaintext: string): Buffer {
  const key = getKey();
  const nonce = randomBytes(NONCE_LENGTH);
  const cipher = createCipheriv(ALGORITHM, key, nonce);

  const encrypted = Buffer.concat([
    cipher.update(plaintext, "utf8"),
    cipher.final(),
  ]);
  const tag = cipher.getAuthTag();

  return Buffer.concat([nonce, encrypted, tag]);
}

/**
 * Decrypt raw bytes produced by encryptToBytes() or Go's gcm.Seal().
 */
export function decryptFromBytes(data: Buffer): string {
  const key = getKey();

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

/**
 * Encrypt a JSON-serializable object to a hex string for Supabase bytea columns.
 * Supabase expects bytea as hex: \\x prefix.
 */
export function encryptCredentials(creds: Record<string, string>): string {
  const plaintext = JSON.stringify(creds);
  const encrypted = encryptToBytes(plaintext);
  return "\\x" + encrypted.toString("hex");
}

/**
 * Decrypt a bytea hex string from Supabase back to a JSON object.
 */
export function decryptCredentials(hex: string): Record<string, string> {
  const clean = hex.startsWith("\\x") ? hex.slice(2) : hex;
  const data = Buffer.from(clean, "hex");
  const plaintext = decryptFromBytes(data);
  return JSON.parse(plaintext);
}
