/**
 * Engagent Cookie Capture — captures all cookies from the current tab's
 * domain and sends them to the admin portal.
 */

const PORTAL_URL = "http://localhost:3000";

// Map domains to provider IDs for the save-cookies endpoint
const DOMAIN_TO_PROVIDER = {
  "linkedin.com": "linkedin",
  "instagram.com": "instagram",
  "x.com": "x",
  "twitter.com": "x",
  "streeteasy.com": "streeteasy",
  "zillow.com": "zillow",
  "instructure.com": "canvas",
};

function getProvider(hostname) {
  for (const [domain, provider] of Object.entries(DOMAIN_TO_PROVIDER)) {
    if (hostname === domain || hostname.endsWith("." + domain)) {
      return provider;
    }
  }
  // Unknown site — use the domain as the provider name
  return hostname.replace(/^www\./, "").split(".")[0];
}

async function init() {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  const domainEl = document.getElementById("domain");
  const btn = document.getElementById("captureBtn");
  const status = document.getElementById("status");

  if (!tab?.url) {
    domainEl.textContent = "No active tab";
    btn.disabled = true;
    return;
  }

  let hostname;
  try {
    hostname = new URL(tab.url).hostname;
  } catch {
    domainEl.textContent = "Invalid URL";
    btn.disabled = true;
    return;
  }

  domainEl.textContent = hostname;

  btn.addEventListener("click", async () => {
    btn.disabled = true;
    btn.textContent = "Capturing...";
    status.textContent = "";
    status.className = "status";

    try {
      // Get all cookies for this domain
      const rootDomain = "." + hostname.split(".").slice(-2).join(".");
      const cookies = await chrome.cookies.getAll({ domain: rootDomain });

      // Also get cookies for the exact hostname
      const exactCookies = await chrome.cookies.getAll({ url: tab.url });
      const seen = new Set(cookies.map((c) => c.name));
      for (const c of exactCookies) {
        if (!seen.has(c.name)) cookies.push(c);
      }

      if (cookies.length === 0) {
        throw new Error("No cookies found. Are you logged in?");
      }

      const cookieMap = {};
      for (const c of cookies) {
        cookieMap[c.name] = c.value;
      }

      const provider = getProvider(hostname);

      const res = await fetch(`${PORTAL_URL}/api/integrations/save-cookies`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          provider,
          cookies: cookieMap,
          label: `${provider} Account`,
        }),
      });

      if (!res.ok) {
        const data = await res.json().catch(() => ({}));
        throw new Error(data.error || `Error ${res.status}`);
      }

      btn.textContent = "Captured!";
      btn.classList.add("success");
      status.textContent = `${cookies.length} cookies saved as "${provider}"`;
      status.className = "status success";

      setTimeout(() => {
        btn.textContent = "Capture Cookies";
        btn.classList.remove("success");
        btn.disabled = false;
      }, 3000);
    } catch (err) {
      btn.textContent = "Failed";
      btn.classList.add("error");
      status.textContent = err.message;
      status.className = "status error";

      setTimeout(() => {
        btn.textContent = "Capture Cookies";
        btn.classList.remove("error");
        btn.disabled = false;
      }, 3000);
    }
  });
}

init();
