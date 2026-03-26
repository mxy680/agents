import { NextRequest, NextResponse } from "next/server"
import { writeFileSync, mkdirSync, readFileSync, existsSync } from "fs"

const RESULTS_DIR = "/tmp/nyc_assemblage"
const RESULTS_FILE = `${RESULTS_DIR}/zillow_results.json`

const CORS_HEADERS = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Methods": "POST, OPTIONS",
  "Access-Control-Allow-Headers": "Content-Type",
}

/**
 * Handle CORS preflight.
 */
export async function OPTIONS() {
  return new NextResponse(null, { status: 204, headers: CORS_HEADERS })
}

/**
 * POST /api/integrations/zillow/scrape-results
 *
 * Receives Zillow listing data from the Chrome extension scraper.
 * Merges into /tmp/nyc_assemblage/zillow_results.json for the pipeline.
 *
 * No auth — localhost-only, called by the Chrome extension which can't
 * send portal cookies (different origin).
 */
export async function POST(request: NextRequest) {
  let listings: unknown[]
  try {
    const body = await request.json()
    listings = body.listings
  } catch {
    return NextResponse.json(
      { error: "Invalid request body" },
      { status: 400, headers: CORS_HEADERS }
    )
  }

  if (!Array.isArray(listings)) {
    return NextResponse.json(
      { error: "listings must be an array" },
      { status: 400, headers: CORS_HEADERS }
    )
  }

  try {
    mkdirSync(RESULTS_DIR, { recursive: true })

    // Merge with existing results (for incremental batching)
    let existing: Record<string, unknown> = {}
    if (existsSync(RESULTS_FILE)) {
      try {
        const raw = JSON.parse(readFileSync(RESULTS_FILE, "utf8"))
        if (Array.isArray(raw)) {
          for (const item of raw) {
            const zpid = (item as Record<string, unknown>)?.zpid
            if (zpid) existing[String(zpid)] = item
          }
        }
      } catch {
        // Corrupt file — start fresh
      }
    }

    // Merge new listings
    for (const item of listings) {
      const zpid = (item as Record<string, unknown>)?.zpid
      if (zpid) existing[String(zpid)] = item
    }

    const merged = Object.values(existing)
    writeFileSync(RESULTS_FILE, JSON.stringify(merged, null, 2))

    return NextResponse.json(
      {
        ok: true,
        new: listings.length,
        total: merged.length,
        path: RESULTS_FILE,
        saved_at: new Date().toISOString(),
      },
      { headers: CORS_HEADERS }
    )
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err)
    return NextResponse.json(
      { ok: false, error: message },
      { status: 500, headers: CORS_HEADERS }
    )
  }
}
