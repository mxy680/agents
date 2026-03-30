/**
 * Emdash Zillow Scraper — background service worker.
 *
 * Navigates a tab through Zillow search pages for each NYC ZIP code,
 * extracts listing data from __NEXT_DATA__ JSON, and POSTs results
 * to the portal API.
 *
 * This works because Chrome extensions use chrome.tabs/chrome.scripting
 * APIs (not CDP), which PerimeterX cannot detect.
 */

const DELAY_BETWEEN_ZIPS_MS = 3000;
const PAGE_WAIT_MS = 6000;

// Progress state (polled by popup)
let scrapeState = {
  running: false,
  done: false,
  processed: 0,
  total: 0,
  results: 0,
  currentBorough: "",
  currentZip: "",
};

async function getPortalURL() {
  const result = await chrome.storage.local.get("portalURL");
  return result.portalURL || "http://localhost:3000";
}

/**
 * Injected into Zillow pages to extract listings from __NEXT_DATA__.
 */
function extractZillowListings() {
  const el = document.getElementById("__NEXT_DATA__");
  if (!el) return null;
  try {
    const data = JSON.parse(el.textContent);
    const results =
      data?.props?.pageProps?.searchPageState?.cat1?.searchResults
        ?.listResults || [];
    return results.map((r) => ({
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
  } catch (e) {
    return null;
  }
}

async function runScrape(zipCodes) {
  const portalURL = await getPortalURL();

  scrapeState = {
    running: true,
    done: false,
    processed: 0,
    total: zipCodes.length,
    results: 0,
    currentBorough: "",
    currentZip: "",
  };

  // Create a background tab on zillow.com to establish session
  const tab = await chrome.tabs.create({
    url: "https://www.zillow.com",
    active: false,
  });

  // Wait for PerimeterX session to initialize
  await new Promise((r) => setTimeout(r, 8000));

  const allResults = {};

  for (const { borough, zip } of zipCodes) {
    scrapeState.processed++;
    scrapeState.currentBorough = borough;
    scrapeState.currentZip = zip;

    // Navigate the tab
    await chrome.tabs.update(tab.id, {
      url: `https://www.zillow.com/homes/${borough}-NY-${zip}_rb/`,
    });

    // Wait for page load
    await new Promise((resolve) => {
      function listener(tabId, info) {
        if (tabId === tab.id && info.status === "complete") {
          chrome.tabs.onUpdated.removeListener(listener);
          resolve();
        }
      }
      chrome.tabs.onUpdated.addListener(listener);
      setTimeout(() => {
        chrome.tabs.onUpdated.removeListener(listener);
        resolve();
      }, 25000);
    });

    // Wait for JS to render
    await new Promise((r) => setTimeout(r, PAGE_WAIT_MS));

    // Extract listings
    try {
      const [result] = await chrome.scripting.executeScript({
        target: { tabId: tab.id },
        func: extractZillowListings,
      });

      const listings = result?.result;
      if (listings && listings.length > 0) {
        // Tag with ZIP and borough
        for (const l of listings) {
          if (l.zpid) {
            l._zip = zip;
            l._borough = borough;
            allResults[l.zpid] = l;
          }
        }
        scrapeState.results = Object.keys(allResults).length;

        // POST this ZIP's results immediately
        try {
          await fetch(`${portalURL}/api/integrations/zillow/scrape-results`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ listings }),
          });
        } catch (e) {
          console.error(`[emdash] POST failed for ${zip}:`, e);
        }

        console.log(
          `[emdash] ${scrapeState.processed}/${zipCodes.length} ${borough} ${zip}: ${listings.length} results (${scrapeState.results} total)`
        );
      } else {
        console.log(
          `[emdash] ${scrapeState.processed}/${zipCodes.length} ${borough} ${zip}: 0 results`
        );
      }
    } catch (e) {
      console.error(`[emdash] ${zip} error:`, e);
    }

    await new Promise((r) => setTimeout(r, DELAY_BETWEEN_ZIPS_MS));
  }

  // Close scraping tab
  try {
    await chrome.tabs.remove(tab.id);
  } catch {}

  const listings = Object.values(allResults);

  scrapeState = {
    running: false,
    done: true,
    processed: zipCodes.length,
    total: zipCodes.length,
    results: listings.length,
    currentBorough: "",
    currentZip: "",
  };

  return { ok: true, count: listings.length };
}

// Message listener
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.action === "start-scrape" && message.zipCodes) {
    runScrape(message.zipCodes)
      .then((result) => sendResponse(result))
      .catch((err) => sendResponse({ ok: false, error: err.message }));
    return true; // Keep channel open for async
  }
  if (message.action === "scrape-status") {
    sendResponse({ ...scrapeState });
    return;
  }
});
