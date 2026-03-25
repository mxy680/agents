const domainEl = document.getElementById("domain");
const statusEl = document.getElementById("status");
const copyBtn = document.getElementById("copy-btn");

let currentDomain = "";

function setStatus(message, type) {
  statusEl.textContent = message;
  statusEl.className = "status " + (type || "info");
}

/**
 * Extract the registrable domain (e.g. "yelp.com") from a hostname
 * so that we also pick up subdomain cookies (e.g. ".yelp.com").
 * Falls back to the full hostname if splitting fails.
 */
function extractDomain(url) {
  try {
    const { hostname } = new URL(url);
    // Use the last two parts for a simple eTLD+1 approximation.
    // For common cases (yelp.com, zillow.com, etc.) this is correct.
    const parts = hostname.split(".");
    if (parts.length >= 2) {
      return parts.slice(-2).join(".");
    }
    return hostname;
  } catch {
    return "";
  }
}

// On load, query the active tab to get the current domain
chrome.tabs.query({ active: true, currentWindow: true }, (tabs) => {
  const tab = tabs[0];
  if (!tab || !tab.url) {
    domainEl.textContent = "No active tab";
    domainEl.className = "domain";
    setStatus("Open a website tab first", "error");
    return;
  }

  const url = tab.url;

  // Skip chrome:// and other non-http pages
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
  setStatus("");
});

copyBtn.addEventListener("click", async () => {
  if (!currentDomain) return;

  copyBtn.disabled = true;
  setStatus("Reading cookies...", "info");

  try {
    // Query cookies for both "domain.com" and ".domain.com" (subdomain cookies)
    const [exactCookies, subdomainCookies] = await Promise.all([
      chrome.cookies.getAll({ domain: currentDomain }),
      chrome.cookies.getAll({ domain: "." + currentDomain }),
    ]);

    // Merge, deduplicating by name (exact match takes precedence)
    const cookieMap = {};

    for (const cookie of subdomainCookies) {
      cookieMap[cookie.name] = cookie.value;
    }
    for (const cookie of exactCookies) {
      cookieMap[cookie.name] = cookie.value;
    }

    const cookieCount = Object.keys(cookieMap).length;

    if (cookieCount === 0) {
      setStatus("No cookies found for " + currentDomain, "error");
      copyBtn.disabled = false;
      return;
    }

    const json = JSON.stringify(cookieMap, null, 2);
    await navigator.clipboard.writeText(json);

    setStatus(
      "Copied " + cookieCount + " cookie" + (cookieCount === 1 ? "" : "s"),
      "success"
    );

    // Re-enable button after short delay
    setTimeout(() => {
      copyBtn.disabled = false;
    }, 1500);
  } catch (err) {
    const message = err instanceof Error ? err.message : String(err);
    setStatus("Error: " + message, "error");
    copyBtn.disabled = false;
  }
});
