import { NextRequest, NextResponse } from "next/server"
import { createClient } from "@/lib/supabase/server"
import { isAdmin } from "@/lib/admin"
import { writeFileSync, mkdirSync } from "fs"

const RESULTS_DIR = "/tmp/nyc_assemblage"
const RESULTS_FILE = `${RESULTS_DIR}/zillow_results.json`

/**
 * POST /api/integrations/zillow/scrape-results
 *
 * Receives Zillow listing data from the Chrome extension scraper.
 * Saves to /tmp/nyc_assemblage/zillow_results.json for the pipeline.
 */
export async function POST(request: NextRequest) {
  const supabase = await createClient()
  const {
    data: { user },
  } = await supabase.auth.getUser()

  if (!user || !isAdmin(user.email)) {
    return NextResponse.json({ error: "Unauthorized" }, { status: 401 })
  }

  let listings: unknown[]
  try {
    const body = await request.json()
    listings = body.listings
  } catch {
    return NextResponse.json({ error: "Invalid request body" }, { status: 400 })
  }

  if (!Array.isArray(listings)) {
    return NextResponse.json(
      { error: "listings must be an array" },
      { status: 400 }
    )
  }

  try {
    mkdirSync(RESULTS_DIR, { recursive: true })
    writeFileSync(RESULTS_FILE, JSON.stringify(listings, null, 2))

    return NextResponse.json({
      ok: true,
      count: listings.length,
      path: RESULTS_FILE,
      saved_at: new Date().toISOString(),
    })
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    return NextResponse.json({ ok: false, error: message }, { status: 500 })
  }
}
