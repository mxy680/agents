import { createClient } from "@supabase/supabase-js"
import { decrypt } from "./lib/crypto"

const supabase = createClient(process.env.NEXT_PUBLIC_SUPABASE_URL!, process.env.SUPABASE_SERVICE_ROLE_KEY!)

const { data, error } = await supabase
  .from("user_integrations")
  .select("provider, label, credentials, status")
  .eq("provider", "linkedin")
  .eq("status", "active")
  .limit(1)

if (error) { console.error(error); process.exit(1) }
if (!data?.length) { console.log("No linkedin integration found"); process.exit(1) }

const row = data[0]
const raw = row.credentials as string
const hex = raw.startsWith("\\x") ? raw.slice(2) : raw
const buf = Buffer.from(hex, "hex")
const decrypted = decrypt(buf)
const creds = JSON.parse(decrypted) as Record<string, string>

console.log("Keys:", Object.keys(creds).join(", "))
for (const [k, v] of Object.entries(creds)) {
  console.log(`${k}: ${v.substring(0, 20)}...`)
}
console.log("=== JSON ===")
console.log(JSON.stringify(creds))
