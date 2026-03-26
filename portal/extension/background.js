/**
 * Background service worker — monitors cookie changes for watched domains
 * and syncs them to the Emdash portal API.
 *
 * Watched domains are configured via the popup. When a cookie changes for
 * a watched domain, we debounce for 2 seconds then push the full cookie
 * jar to /api/integrations/cookies/sync.
 */

const SYNC_DEBOUNCE_MS = 2000;
const SYNC_INTERVAL_MINUTES = 1; // Also sync every minute as a heartbeat

// Domains we watch and their provider names
const DOMAIN_PROVIDER_MAP = {
  "zillow.com": "zillow",
  "streeteasy.com": "streeteasy",
};

// Debounce timers per domain
const pendingSyncs = {};

/**
 * Get the portal URL from storage, or default to localhost.
 */
async function getPortalURL() {
  const result = await chrome.storage.local.get("portalURL");
  return result.portalURL || "http://localhost:3000";
}

/**
 * Get all cookies for a domain (including subdomain cookies).
 */
async function getCookiesForDomain(domain) {
  const [exact, sub] = await Promise.all([
    chrome.cookies.getAll({ domain }),
    chrome.cookies.getAll({ domain: "." + domain }),
  ]);

  const cookieMap = {};
  for (const c of sub) cookieMap[c.name] = c.value;
  for (const c of exact) cookieMap[c.name] = c.value;

  return Object.entries(cookieMap)
    .filter(([, v]) => v)
    .map(([k, v]) => `${k}=${v}`)
    .join("; ");
}

/**
 * Push cookies to the portal API.
 */
async function syncCookies(domain) {
  const provider = DOMAIN_PROVIDER_MAP[domain];
  if (!provider) return;

  const portalURL = await getPortalURL();
  const cookieString = await getCookiesForDomain(domain);

  if (!cookieString) {
    console.log(`[emdash] No cookies for ${domain}, skipping sync`);
    return;
  }

  try {
    const resp = await fetch(`${portalURL}/api/integrations/cookies/sync`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ provider, cookies: cookieString }),
      credentials: "include", // Send portal auth cookies
    });

    if (resp.ok) {
      const data = await resp.json();
      console.log(`[emdash] Synced ${provider} cookies (${cookieString.length} chars)`);

      // Update badge
      chrome.action.setBadgeText({ text: "✓" });
      chrome.action.setBadgeBackgroundColor({ color: "#22c55e" });
      setTimeout(() => chrome.action.setBadgeText({ text: "" }), 3000);
    } else {
      console.error(`[emdash] Sync failed for ${provider}: HTTP ${resp.status}`);
      chrome.action.setBadgeText({ text: "!" });
      chrome.action.setBadgeBackgroundColor({ color: "#ef4444" });
    }
  } catch (err) {
    console.error(`[emdash] Sync error for ${provider}:`, err.message);
  }
}

/**
 * Debounced sync — waits SYNC_DEBOUNCE_MS after the last cookie change
 * before syncing, to batch rapid cookie rotations.
 */
function debouncedSync(domain) {
  if (pendingSyncs[domain]) {
    clearTimeout(pendingSyncs[domain]);
  }
  pendingSyncs[domain] = setTimeout(() => {
    delete pendingSyncs[domain];
    syncCookies(domain);
  }, SYNC_DEBOUNCE_MS);
}

/**
 * Check if a cookie change is for a watched domain.
 */
function getWatchedDomain(cookieDomain) {
  // Strip leading dot
  const domain = cookieDomain.startsWith(".")
    ? cookieDomain.slice(1)
    : cookieDomain;

  for (const watched of Object.keys(DOMAIN_PROVIDER_MAP)) {
    if (domain === watched || domain.endsWith("." + watched)) {
      return watched;
    }
  }
  return null;
}

// Listen for cookie changes
chrome.cookies.onChanged.addListener((changeInfo) => {
  const watchedDomain = getWatchedDomain(changeInfo.cookie.domain);
  if (watchedDomain) {
    debouncedSync(watchedDomain);
  }
});

// Periodic heartbeat sync — ensures cookies are fresh even without changes
chrome.alarms.create("cookie-sync-heartbeat", {
  periodInMinutes: SYNC_INTERVAL_MINUTES,
});

chrome.alarms.onAlarm.addListener((alarm) => {
  if (alarm.name === "cookie-sync-heartbeat") {
    for (const domain of Object.keys(DOMAIN_PROVIDER_MAP)) {
      syncCookies(domain);
    }
  }
});

// On install/update, do an initial sync
chrome.runtime.onInstalled.addListener(() => {
  console.log("[emdash] Extension installed/updated — syncing all cookies");
  for (const domain of Object.keys(DOMAIN_PROVIDER_MAP)) {
    syncCookies(domain);
  }
});

// ============================================================
// Zillow Scraper — triggered by portal via external message
// ============================================================

const SCRAPE_DELAY_MS = 3000;
const SCRAPE_PAGE_WAIT_MS = 6000;

// Scrape progress tracking
let scrapeState = { running: false, done: false, processed: 0, total: 0, results: 0, currentBorough: "", currentZip: "" };

/**
 * Content script function injected into Zillow pages to extract listings.
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

/**
 * Listen for messages from the portal to start a scrape job.
 * The portal sends: { action: "start-scrape", zipCodes: [{borough, zip}] }
 */
chrome.runtime.onMessageExternal.addListener(
  (message, sender, sendResponse) => {
    if (message.action === "start-scrape" && message.zipCodes) {
      console.log(
        `[emdash] Starting scrape job: ${message.zipCodes.length} ZIP codes`
      );
      runScrape(message.zipCodes)
        .then((result) => sendResponse(result))
        .catch((err) => sendResponse({ ok: false, error: err.message }));
      return true; // Keep message channel open for async response
    }

    if (message.action === "ping") {
      sendResponse({ ok: true, version: "2.1.0" });
      return;
    }
  }
);

/**
 * Also listen for internal messages (from popup).
 */
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.action === "start-scrape" && message.zipCodes) {
    runScrape(message.zipCodes)
      .then((result) => sendResponse(result))
      .catch((err) => sendResponse({ ok: false, error: err.message }));
    return true;
  }
  if (message.action === "scrape-status") {
    sendResponse({ ...scrapeState });
    return;
  }
});

async function runScrape(zipCodes) {
  const portalURL = await getPortalURL();

  scrapeState = { running: true, done: false, processed: 0, total: zipCodes.length, results: 0, currentBorough: "", currentZip: "" };

  // Create a tab for scraping (inactive)
  const tab = await chrome.tabs.create({
    url: "https://www.zillow.com",
    active: false,
  });

  // Wait for initial Zillow load (PerimeterX session init)
  await new Promise((r) => setTimeout(r, 8000));

  const allResults = {};
  let processed = 0;

  for (const { borough, zip } of zipCodes) {
    processed++;
    scrapeState.processed = processed;
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

    // Wait for JS rendering
    await new Promise((r) => setTimeout(r, SCRAPE_PAGE_WAIT_MS));

    // Extract listings
    try {
      const [result] = await chrome.scripting.executeScript({
        target: { tabId: tab.id },
        func: extractZillowListings,
      });

      const listings = result?.result;
      if (listings && listings.length > 0) {
        for (const l of listings) {
          if (l.zpid) {
            l._zip = zip;
            l._borough = borough;
            allResults[l.zpid] = l;
          }
        }
        scrapeState.results = Object.keys(allResults).length;
        console.log(`[emdash] ${processed}/${zipCodes.length} ${borough} ${zip}: ${listings.length} results (${scrapeState.results} total)`);
      } else {
        console.log(`[emdash] ${processed}/${zipCodes.length} ${borough} ${zip}: 0 results`);
      }
    } catch (e) {
      console.error(`[emdash] ${zip} script error:`, e);
    }

    // Rate limit
    await new Promise((r) => setTimeout(r, SCRAPE_DELAY_MS));
  }

  // Close scraping tab
  try {
    await chrome.tabs.remove(tab.id);
  } catch {}

  // Post results to portal
  const listings = Object.values(allResults);
  try {
    await fetch(`${portalURL}/api/integrations/zillow/scrape-results`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ listings }),
      credentials: "include",
    });
  } catch (e) {
    console.error("[emdash] Failed to post results:", e);
  }

  // Also save to a file the pipeline can read
  scrapeState = { running: false, done: true, processed: zipCodes.length, total: zipCodes.length, results: listings.length, currentBorough: "", currentZip: "" };

  console.log(`[emdash] Scrape complete: ${listings.length} unique listings`);
  return { ok: true, count: listings.length };
}
