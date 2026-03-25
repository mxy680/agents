/**
 * Zillow search scraper — runs inside the Chrome extension context.
 *
 * Called by the portal job runner via the extension's message API.
 * Opens Zillow search URLs in a tab, extracts listing data from the
 * page's embedded JSON, and posts results back to the portal.
 *
 * This works because it runs inside Chrome's normal context — no
 * debugger attached, no Playwright, no CDP. PerimeterX can't detect it.
 */

const DELAY_BETWEEN_ZIPS_MS = 3000;
const PAGE_LOAD_WAIT_MS = 5000;

/**
 * Extract listing data from a Zillow search results page.
 * Called via chrome.scripting.executeScript in the target tab.
 */
function extractListings() {
  // Zillow embeds search results in a <script id="__NEXT_DATA__"> tag
  const nextDataEl = document.getElementById("__NEXT_DATA__");
  if (nextDataEl) {
    try {
      const data = JSON.parse(nextDataEl.textContent);
      const searchResults =
        data?.props?.pageProps?.searchPageState?.cat1?.searchResults?.listResults;
      if (searchResults) {
        return searchResults.map((r) => ({
          zpid: r.zpid || r.id,
          address: r.address || r.statusText,
          price: r.unformattedPrice || r.price,
          beds: r.beds,
          baths: r.baths,
          sqft: r.area,
          homeType: r.homeType,
          status: r.statusType || r.homeStatus,
          zillowUrl: r.detailUrl
            ? "https://www.zillow.com" + r.detailUrl
            : undefined,
          latitude: r.latLong?.latitude,
          longitude: r.latLong?.longitude,
          daysOnMarket: r.variableData?.text
            ? parseInt(r.variableData.text)
            : undefined,
        }));
      }
    } catch (e) {
      console.error("[emdash] Failed to parse __NEXT_DATA__:", e);
    }
  }

  // Fallback: try extracting from window.__PRELOADED_STATE__
  if (window.__PRELOADED_STATE__) {
    try {
      const state = window.__PRELOADED_STATE__;
      // Navigate to search results in the state tree
      const results = state?.searchPage?.queryState?.mapResults ||
        state?.searchPage?.searchResults?.listResults || [];
      return results.map((r) => ({
        zpid: r.zpid,
        address: r.address,
        price: r.price,
        latitude: r.lat || r.latitude,
        longitude: r.lng || r.longitude,
      }));
    } catch (e) {
      console.error("[emdash] Failed to parse __PRELOADED_STATE__:", e);
    }
  }

  return null;
}

/**
 * Scrape a single ZIP code by navigating a tab to the Zillow search URL.
 */
async function scrapeZip(tabId, borough, zipCode) {
  const url = `https://www.zillow.com/homes/${borough}-NY-${zipCode}_rb/`;

  // Navigate the tab
  await chrome.tabs.update(tabId, { url });

  // Wait for page to load
  await new Promise((resolve) => {
    function onComplete(updatedTabId, changeInfo) {
      if (updatedTabId === tabId && changeInfo.status === "complete") {
        chrome.tabs.onUpdated.removeListener(onComplete);
        resolve();
      }
    }
    chrome.tabs.onUpdated.addListener(onComplete);
    // Timeout after 30s
    setTimeout(() => {
      chrome.tabs.onUpdated.removeListener(onComplete);
      resolve();
    }, 30000);
  });

  // Wait extra for JS to render
  await new Promise((r) => setTimeout(r, PAGE_LOAD_WAIT_MS));

  // Extract listings from the page
  try {
    const results = await chrome.scripting.executeScript({
      target: { tabId },
      func: extractListings,
    });

    if (results?.[0]?.result) {
      return results[0].result;
    }
  } catch (e) {
    console.error(`[emdash] Script execution failed for ${zipCode}:`, e);
  }

  return null;
}

/**
 * Run the full Zillow scrape job.
 * Called from background.js when it receives a "start-scrape" message.
 */
async function runScrapeJob(portalURL, zipCodes) {
  // Create a tab for scraping
  const tab = await chrome.tabs.create({ url: "https://www.zillow.com", active: false });

  // Wait for initial page load (establishes PerimeterX session)
  await new Promise((r) => setTimeout(r, 8000));

  const allResults = {};
  let processed = 0;
  const total = zipCodes.length;

  for (const { borough, zip } of zipCodes) {
    processed++;
    console.log(`[emdash] Scraping ${processed}/${total}: ${borough} ${zip}`);

    try {
      const listings = await scrapeZip(tab.id, borough, zip);

      if (listings && listings.length > 0) {
        for (const listing of listings) {
          if (listing.zpid) {
            listing._zip = zip;
            listing._borough = borough;
            allResults[listing.zpid] = listing;
          }
        }
        console.log(`[emdash]   ${listings.length} results`);
      } else {
        console.log(`[emdash]   0 results (or blocked)`);
      }

      // Post progress to portal
      if (processed % 10 === 0) {
        try {
          await fetch(`${portalURL}/api/integrations/zillow/scrape-progress`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              processed,
              total,
              results_so_far: Object.keys(allResults).length,
            }),
            credentials: "include",
          });
        } catch {
          // Ignore progress reporting errors
        }
      }
    } catch (e) {
      console.error(`[emdash] Error scraping ${zip}:`, e);
    }

    // Rate limit
    await new Promise((r) => setTimeout(r, DELAY_BETWEEN_ZIPS_MS));
  }

  // Close the scraping tab
  try {
    await chrome.tabs.remove(tab.id);
  } catch {
    // Tab may already be closed
  }

  // Post final results to portal
  const resultsList = Object.values(allResults);
  try {
    const resp = await fetch(`${portalURL}/api/integrations/zillow/scrape-results`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ listings: resultsList }),
      credentials: "include",
    });

    if (resp.ok) {
      console.log(`[emdash] Scrape complete. ${resultsList.length} unique listings posted.`);
      return { ok: true, count: resultsList.length };
    } else {
      console.error(`[emdash] Failed to post results: HTTP ${resp.status}`);
      return { ok: false, error: `HTTP ${resp.status}` };
    }
  } catch (e) {
    console.error("[emdash] Failed to post results:", e);
    return { ok: false, error: e.message };
  }
}

// Export for use in background.js
if (typeof globalThis !== "undefined") {
  globalThis.runScrapeJob = runScrapeJob;
}
