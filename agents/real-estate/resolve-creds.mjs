#!/usr/bin/env node
/**
 * Resolves integration credentials from Supabase and prints them as
 * shell export statements. Used by run-local.sh to inject fresh
 * credentials into the agent's environment.
 *
 * Usage: eval "$(node resolve-creds.mjs)"
 */

// Resolve @supabase/supabase-js — try multiple locations
import { fileURLToPath } from "url"
import { dirname, join } from "path"
const __dirname = dirname(fileURLToPath(import.meta.url))

let createClient;
for (const base of [
  join(__dirname, "..", "..", "portal", "node_modules"),  // local dev
  join("/app", "node_modules"),                            // Fly.io container
]) {
  try {
    const mod = await import(join(base, "@supabase", "supabase-js", "dist", "index.mjs"));
    createClient = mod.createClient;
    break;
  } catch {}
}
if (!createClient) {
  const mod = await import("@supabase/supabase-js");
  createClient = mod.createClient;
}

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
    case "github":
      if (creds.access_token) env.GITHUB_ACCESS_TOKEN = creds.access_token
      if (creds.refresh_token) env.GITHUB_REFRESH_TOKEN = creds.refresh_token
      break
    case "zillow":
      if (creds.all_cookies) env.ZILLOW_COOKIES = creds.all_cookies
      if (creds.proxy_url) env.ZILLOW_PROXY_URL = creds.proxy_url
      break
    case "vercel":
      if (creds.token) env.VERCEL_TOKEN = creds.token
      if (creds.team_id) env.VERCEL_TEAM_ID = creds.team_id
      break
    case "cloudflare":
      if (creds.token) env.CLOUDFLARE_API_TOKEN = creds.token
      if (creds.account_id) env.CLOUDFLARE_ACCOUNT_ID = creds.account_id
      break
    case "linear":
      if (creds.token) env.LINEAR_API_KEY = creds.token
      break
    case "fly":
      if (creds.token) env.FLY_API_TOKEN = creds.token
      break
    case "gcp":
      if (creds.token) env.GCP_ACCESS_TOKEN = creds.token
      if (creds.service_account_json) env.GCP_SERVICE_ACCOUNT_JSON = creds.service_account_json
      if (creds.project_id) env.GCP_PROJECT_ID = creds.project_id
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

// Pass through ScraperAPI key as ZILLOW_PROXY_URL if set in environment
// (from Doppler) and not already emitted from creds.
if (process.env.SCRAPERAPI_KEY && !process.env.ZILLOW_PROXY_URL) {
  const key = process.env.SCRAPERAPI_KEY
  // ultra_premium=true required for Zillow (PerimeterX-protected site)
  const proxyUrl = `http://scraperapi.ultra_premium=true:${key}@proxy-server.scraperapi.com:8001`
  console.log(`export ZILLOW_PROXY_URL='${proxyUrl}'`)
  emittedCount++
}

// Warn if no credentials were resolved, but don't fail — some jobs use only public APIs
if (emittedCount === 0) {
  process.stderr.write(`WARN: No credentials resolved from Supabase. Jobs using public APIs will still work.\n`)
  // Emit a dummy export so the sourcing shell script sees a non-empty file
  console.log(`export __CREDS_RESOLVED=0`)
}
