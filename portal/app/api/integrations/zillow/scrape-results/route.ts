import { NextRequest, NextResponse } from "next/server"
import { createAdminClient } from "@/lib/supabase/admin"

const CORS_HEADERS = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Methods": "POST, OPTIONS",
  "Access-Control-Allow-Headers": "Content-Type",
}

export async function OPTIONS() {
  return new NextResponse(null, { status: 204, headers: CORS_HEADERS })
}

interface Listing {
  zpid: string
  _borough?: string
  _zip?: string
  [key: string]: unknown
}

/**
 * POST /api/integrations/zillow/scrape-results
 *
 * Receives Zillow listing data from the Chrome extension scraper.
 * Upserts into zillow_scrape_listings table in Supabase.
 */
export async function POST(request: NextRequest) {
  let listings: Listing[]
  try {
    const body = await request.json()
    listings = body.listings
  } catch {
    return NextResponse.json(
      { error: "Invalid request body" },
      { status: 400, headers: CORS_HEADERS }
    )
  }

  if (!Array.isArray(listings) || listings.length === 0) {
    return NextResponse.json(
      { error: "listings must be a non-empty array" },
      { status: 400, headers: CORS_HEADERS }
    )
  }

  const batchId = new Date().toISOString().slice(0, 10) // e.g. "2026-03-25"

  try {
    const admin = createAdminClient()

    const rows = listings
      .filter((l) => l.zpid)
      .map((l) => ({
        zpid: String(l.zpid),
        data: l,
        borough: l._borough || null,
        zip: l._zip || null,
        scraped_at: new Date().toISOString(),
        scrape_batch: batchId,
      }))

    // Upsert in chunks of 500
    let upserted = 0
    for (let i = 0; i < rows.length; i += 500) {
      const chunk = rows.slice(i, i + 500)
      const { error } = await admin
        .from("zillow_scrape_listings")
        .upsert(chunk, { onConflict: "zpid" })

      if (error) {
        console.error("[scrape-results] Upsert error:", error.message)
        return NextResponse.json(
          { ok: false, error: error.message },
          { status: 500, headers: CORS_HEADERS }
        )
      }
      upserted += chunk.length
    }

    // Get total count
    const { count } = await admin
      .from("zillow_scrape_listings")
      .select("zpid", { count: "exact", head: true })
      .eq("scrape_batch", batchId)

    return NextResponse.json(
      {
        ok: true,
        new: upserted,
        total: count ?? upserted,
        batch: batchId,
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
