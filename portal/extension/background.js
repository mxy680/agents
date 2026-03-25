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
