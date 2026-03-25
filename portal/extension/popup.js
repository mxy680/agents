const domainEl = document.getElementById("domain");
const statusEl = document.getElementById("status");
const copyBtn = document.getElementById("copy-btn");
const syncBtn = document.getElementById("sync-btn");
const portalUrlInput = document.getElementById("portal-url");

let currentDomain = "";

function setStatus(message, type) {
  statusEl.textContent = message;
  statusEl.className = "status " + (type || "info");
}

/**
 * Extract the registrable domain (e.g. "zillow.com") from a hostname.
 */
function extractDomain(url) {
  try {
    const { hostname } = new URL(url);
    const parts = hostname.split(".");
    if (parts.length >= 2) {
      return parts.slice(-2).join(".");
    }
    return hostname;
  } catch {
    return "";
  }
}

// Load portal URL from storage
chrome.storage.local.get("portalURL", (result) => {
  portalUrlInput.value = result.portalURL || "http://localhost:3000";
});

// Save portal URL on change
portalUrlInput.addEventListener("change", () => {
  const url = portalUrlInput.value.trim().replace(/\/+$/, "");
  chrome.storage.local.set({ portalURL: url });
  setStatus("Portal URL saved", "success");
  setTimeout(() => setStatus(""), 2000);
});

// On load, query the active tab
chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
  const tab = tabs[0];
  if (!tab || !tab.url) {
    domainEl.textContent = "No active tab";
    domainEl.className = "domain";
    setStatus("Open a website tab first", "error");
    return;
  }

  const url = tab.url;
  if (!url.startsWith("http://") && !url.startsWith("https://")) {
    domainEl.textContent = "Not a web page";
    domainEl.className = "domain";
    setStatus("Navigate to a website first", "error");
    return;
  }

  currentDomain = extractDomain(url);
  domainEl.textContent = currentDomain;
  domainEl.className = "domain";
  copyBtn.disabled = false;
  syncBtn.disabled = false;
  setStatus("");
});

// Copy cookies to clipboard
copyBtn.addEventListener("click", async () => {
  if (!currentDomain) return;
  copyBtn.disabled = true;
  setStatus("Reading cookies...", "info");

  try {
    const [exactCookies, subdomainCookies] = await Promise.all([
      chrome.cookies.getAll({ domain: currentDomain }),
      chrome.cookies.getAll({ domain: "." + currentDomain }),
    ]);

    const cookieMap = {};
    for (const cookie of subdomainCookies) cookieMap[cookie.name] = cookie.value;
    for (const cookie of exactCookies) cookieMap[cookie.name] = cookie.value;

    const cookieCount = Object.keys(cookieMap).length;
    if (cookieCount === 0) {
      setStatus("No cookies found for " + currentDomain, "error");
      copyBtn.disabled = false;
      return;
    }

    const json = JSON.stringify(cookieMap, null, 2);
    await navigator.clipboard.writeText(json);

    setStatus("Copied " + cookieCount + " cookies", "success");
    setTimeout(() => {
      copyBtn.disabled = false;
      setStatus("");
    }, 1500);
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    setStatus("Error: " + message, "error");
    copyBtn.disabled = false;
  }
});

// Manual sync for current domain
syncBtn.addEventListener("click", async () => {
  if (!currentDomain) return;
  syncBtn.disabled = true;
  setStatus("Syncing...", "info");

  const DOMAIN_PROVIDER_MAP = {
    "zillow.com": "zillow",
    "streeteasy.com": "streeteasy",
  };

  const provider = DOMAIN_PROVIDER_MAP[currentDomain];
  if (!provider) {
    setStatus("Not a synced provider: " + currentDomain, "error");
    syncBtn.disabled = false;
    return;
  }

  try {
    const [exact, sub] = await Promise.all([
      chrome.cookies.getAll({ domain: currentDomain }),
      chrome.cookies.getAll({ domain: "." + currentDomain }),
    ]);
    const cookieMap = {};
    for (const c of sub) cookieMap[c.name] = c.value;
    for (const c of exact) cookieMap[c.name] = c.value;
    const cookieString = Object.entries(cookieMap)
      .filter(([, v]) => v)
      .map(([k, v]) => `${k}=${v}`)
      .join("; ");

    const portalURL = portalUrlInput.value.trim().replace(/\/+$/, "");
    const resp = await fetch(`${portalURL}/api/integrations/cookies/sync`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ provider, cookies: cookieString }),
      credentials: "include",
    });

    if (resp.ok) {
      setStatus("Synced " + provider + " cookies", "success");
      updateProviderStatus(provider, true);
    } else {
      setStatus("Sync failed: HTTP " + resp.status, "error");
      updateProviderStatus(provider, false);
    }
  } catch (err) {
    setStatus("Sync error: " + err.message, "error");
  }

  setTimeout(() => {
    syncBtn.disabled = false;
  }, 1500);
});

function updateProviderStatus(provider, ok) {
  const el = document.getElementById(provider + "-status");
  if (!el) return;
  if (ok) {
    el.textContent = "synced";
    el.className = "active";
  } else {
    el.textContent = "error";
    el.className = "inactive";
  }
}
