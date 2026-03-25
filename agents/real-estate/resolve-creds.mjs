#!/usr/bin/env node
/**
 * Resolves integration credentials from Supabase and prints them as
 * shell export statements. Used by run-local.sh to inject fresh
 * credentials into the agent's environment.
 *
 * Usage: eval "$(node resolve-creds.mjs)"
 */

import { createClient } from "@supabase/supabase-js"

// Fail fast if required env vars are missing
for (const key of ["ENCRYPTION_MASTER_KEY", "NEXT_PUBLIC_SUPABASE_URL", "SUPABASE_SERVICE_ROLE_KEY"]) {
  if (!process.env[key]) {
    process.stderr.write(`ERROR: ${key} is not set. Check Doppler config.\n`)
    process.exit(1)
  }
}
import crypto from "crypto"

function decrypt(buf) {
  const key = Buffer.from(process.env.ENCRYPTION_MASTER_KEY, "hex")
  const nonce = buf.subarray(0, 12)
  const tag = buf.subarray(buf.length - 16)
  const ciphertext = buf.subarray(12, buf.length - 16)
  const decipher = crypto.createDecipheriv("aes-256-gcm", key, nonce)
  decipher.setAuthTag(tag)
  return Buffer.concat([decipher.update(ciphertext), decipher.final()]).toString("utf8")
}

function credentialsToEnv(provider, creds) {
  const env = {}
  switch (provider) {
    case "google":
      if (creds.access_token) env.GOOGLE_ACCESS_TOKEN = creds.access_token
      if (creds.refresh_token) env.GOOGLE_REFRESH_TOKEN = creds.refresh_token
      break
    case "zillow":
      if (creds.all_cookies) env.ZILLOW_COOKIES = creds.all_cookies
      if (creds.proxy_url) env.ZILLOW_PROXY_URL = creds.proxy_url
      break
    case "streeteasy":
      if (creds.all_cookies) env.STREETEASY_COOKIES = creds.all_cookies
      break
  }
  return env
}

const supabase = createClient(
  process.env.NEXT_PUBLIC_SUPABASE_URL,
  process.env.SUPABASE_SERVICE_ROLE_KEY
)

const { data: integrations, error: queryError } = await supabase
  .from("user_integrations")
  .select("provider, credentials")
  .eq("status", "active")

// #2: Check for Supabase query errors
if (queryError) {
  process.stderr.write(`ERROR: Supabase query failed: ${queryError.message}\n`)
  process.exit(1)
}

let emittedCount = 0
for (const row of integrations ?? []) {
  try {
    const hex = row.credentials.startsWith("\\x") ? row.credentials.slice(2) : row.credentials
    const decrypted = decrypt(Buffer.from(hex, "hex"))
    const creds = JSON.parse(decrypted)
    const env = credentialsToEnv(row.provider, creds)
    for (const [k, v] of Object.entries(env)) {
      // Escape single quotes in values for shell safety
      const escaped = v.replace(/'/g, "'\\''")
      console.log(`export ${k}='${escaped}'`)
      emittedCount++
    }
  } catch {
    // Skip providers that fail to decrypt
  }
}

// #2: Warn and exit non-zero if no credentials were resolved
if (emittedCount === 0) {
  process.stderr.write(`ERROR: No credentials resolved from Supabase. Check that user_integrations rows exist with status=active.\n`)
  process.exit(1)
}
