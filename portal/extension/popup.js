const domainEl = document.getElementById("domain");
const statusEl = document.getElementById("status");
const copyBtn = document.getElementById("copy-btn");
const syncBtn = document.getElementById("sync-btn");
const scrapeBtn = document.getElementById("scrape-btn");
const portalUrlInput = document.getElementById("portal-url");
const scrapeProgress = document.getElementById("scrape-progress");
const scrapeBar = document.getElementById("scrape-bar");
const scrapeZips = document.getElementById("scrape-zips");
const scrapeListings = document.getElementById("scrape-listings");
const scrapeStatusText = document.getElementById("scrape-status-text");

let currentDomain = "";

// All NYC ZIP codes for scraping
const NYC_ZIP_CODES = [
  // Bronx
  ...["10451","10452","10453","10454","10455","10456","10457","10458","10459","10460",
      "10461","10462","10463","10464","10465","10466","10467","10468","10469","10470",
      "10471","10472","10473","10474","10475"].map(z => ({ borough: "Bronx", zip: z })),
  // Brooklyn
  ...["11201","11203","11204","11205","11206","11207","11208","11209","11210","11211",
      "11212","11213","11214","11215","11216","11217","11218","11219","11220","11221",
      "11222","11223","11224","11225","11226","11228","11229","11230","11231","11232",
      "11233","11234","11235","11236","11237","11238","11239"].map(z => ({ borough: "Brooklyn", zip: z })),
  // Manhattan
  ...["10001","10002","10003","10009","10010","10011","10012","10013","10014","10016",
      "10019","10021","10022","10023","10024","10025","10026","10027","10028","10029",
      "10030","10031","10032","10033","10034","10035","10037","10039","10040","10128"].map(z => ({ borough: "Manhattan", zip: z })),
  // Queens
  ...["11101","11102","11103","11104","11105","11106","11109","11354","11355","11356",
      "11357","11358","11359","11360","11361","11362","11363","11364","11365","11366",
      "11367","11368","11369","11370","11372","11373","11374","11375","11377","11378",
      "11379","11385","11411","11412","11413","11414","11415","11416","11417","11418",
      "11419","11420","11421","11422","11423"].map(z => ({ borough: "Queens", zip: z })),
];

function setStatus(message, type) {
  statusEl.textContent = message;
  statusEl.className = "status " + (type || "info");
}

function extractDomain(url) {
  try {
    const { hostname } = new URL(url);
    const parts = hostname.split(".");
    if (parts.length >= 2) return parts.slice(-2).join(".");
    return hostname;
  } catch { return ""; }
}

// Load portal URL from storage
chrome.storage.local.get("portalURL", (result) => {
  portalUrlInput.value = result.portalURL || "http://localhost:3000";
});

portalUrlInput.addEventListener("change", () => {
  const url = portalUrlInput.value.trim().replace(/\/+$/, "");
  chrome.storage.local.set({ portalURL: url });
  setStatus("Portal URL saved", "success");
  setTimeout(() => setStatus(""), 2000);
});

// On load, query active tab
chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
  const tab = tabs[0];
  if (!tab || !tab.url || (!tab.url.startsWith("http://") && !tab.url.startsWith("https://"))) {
    domainEl.textContent = tab?.url ? "Not a web page" : "No active tab";
    domainEl.className = "domain";
    return;
  }
  currentDomain = extractDomain(tab.url);
  domainEl.textContent = currentDomain;
  domainEl.className = "domain";
  copyBtn.disabled = false;
  syncBtn.disabled = false;
});

// Copy cookies
copyBtn.addEventListener("click", async () => {
  if (!currentDomain) return;
  copyBtn.disabled = true;
  setStatus("Reading cookies...", "info");
  try {
    const [exact, sub] = await Promise.all([
      chrome.cookies.getAll({ domain: currentDomain }),
      chrome.cookies.getAll({ domain: "." + currentDomain }),
    ]);
    const map = {};
    for (const c of sub) map[c.name] = c.value;
    for (const c of exact) map[c.name] = c.value;
    const count = Object.keys(map).length;
    if (count === 0) { setStatus("No cookies found", "error"); copyBtn.disabled = false; return; }
    await navigator.clipboard.writeText(JSON.stringify(map, null, 2));
    setStatus(`Copied ${count} cookies`, "success");
    setTimeout(() => { copyBtn.disabled = false; setStatus(""); }, 1500);
  } catch (err) {
    setStatus("Error: " + err.message, "error");
    copyBtn.disabled = false;
  }
});

// Sync cookies
syncBtn.addEventListener("click", async () => {
  if (!currentDomain) return;
  syncBtn.disabled = true;
  setStatus("Syncing...", "info");
  const providers = { "zillow.com": "zillow", "streeteasy.com": "streeteasy" };
  const provider = providers[currentDomain];
  if (!provider) { setStatus("Not a synced provider", "error"); syncBtn.disabled = false; return; }
  try {
    const [exact, sub] = await Promise.all([
      chrome.cookies.getAll({ domain: currentDomain }),
      chrome.cookies.getAll({ domain: "." + currentDomain }),
    ]);
    const map = {};
    for (const c of sub) map[c.name] = c.value;
    for (const c of exact) map[c.name] = c.value;
    const cookieString = Object.entries(map).filter(([,v]) => v).map(([k,v]) => `${k}=${v}`).join("; ");
    const portalURL = portalUrlInput.value.trim().replace(/\/+$/, "");
    const resp = await fetch(`${portalURL}/api/integrations/cookies/sync`, {
      method: "POST", headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ provider, cookies: cookieString }), credentials: "include",
    });
    if (resp.ok) {
      setStatus(`Synced ${provider} cookies`, "success");
      updateProviderStatus(provider, true);
    } else {
      setStatus(`Sync failed: HTTP ${resp.status}`, "error");
    }
  } catch (err) { setStatus("Sync error: " + err.message, "error"); }
  setTimeout(() => { syncBtn.disabled = false; }, 1500);
});

// ============================================================
// Zillow Scraper
// ============================================================

scrapeBtn.addEventListener("click", () => {
  scrapeBtn.disabled = true;
  scrapeBtn.textContent = "Scraping...";
  scrapeProgress.classList.add("visible");
  scrapeStatusText.textContent = "Starting...";
  scrapeBar.style.width = "0%";
  scrapeZips.textContent = `0/${NYC_ZIP_CODES.length}`;
  scrapeListings.textContent = "0";

  // Send message to background worker
  chrome.runtime.sendMessage(
    { action: "start-scrape", zipCodes: NYC_ZIP_CODES },
    (response) => {
      if (response?.ok) {
        scrapeStatusText.textContent = "Complete!";
        scrapeBar.style.width = "100%";
        scrapeListings.textContent = String(response.count || 0);
        setStatus(`Scrape done: ${response.count} listings`, "success");
      } else {
        scrapeStatusText.textContent = "Failed";
        setStatus("Scrape failed: " + (response?.error || "unknown"), "error");
      }
      scrapeBtn.disabled = false;
      scrapeBtn.textContent = "Scrape All NYC ZIP Codes";
    }
  );

  // Poll for progress updates from background
  const progressInterval = setInterval(async () => {
    try {
      const status = await chrome.runtime.sendMessage({ action: "scrape-status" });
      if (status?.running) {
        const pct = Math.round((status.processed / status.total) * 100);
        scrapeBar.style.width = `${pct}%`;
        scrapeZips.textContent = `${status.processed}/${status.total}`;
        scrapeListings.textContent = String(status.results);
        scrapeStatusText.textContent = `${status.currentBorough} ${status.currentZip}...`;
      } else if (status?.done) {
        clearInterval(progressInterval);
      }
    } catch {
      // Extension context may be invalidated
      clearInterval(progressInterval);
    }
  }, 1000);
});

function updateProviderStatus(provider, ok) {
  const el = document.getElementById(provider + "-status");
  if (!el) return;
  el.textContent = ok ? "synced" : "error";
  el.className = ok ? "active" : "inactive";
}
