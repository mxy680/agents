/**
 * Engagent Cookie Capture Extension
 *
 * Captures session cookies from supported sites and sends them
 * to the Engagent admin portal for encrypted storage.
 */

const PROVIDERS = [
  { id: "linkedin", name: "LinkedIn", domain: ".linkedin.com" },
  { id: "instagram", name: "Instagram", domain: ".instagram.com" },
  { id: "x", name: "X (Twitter)", domain: ".x.com", altDomains: [".twitter.com"] },
  { id: "streeteasy", name: "StreetEasy", domain: ".streeteasy.com" },
  { id: "canvas", name: "Canvas LMS", domain: ".instructure.com" },
];

function getPortalUrl() {
  return document.getElementById("portalUrl").value.replace(/\/+$/, "");
}

async function getCurrentTabDomain() {
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true });
  if (!tab?.url) return null;
  try {
    return new URL(tab.url).hostname;
  } catch {
    return null;
  }
}

function matchesDomain(hostname, provider) {
  const domains = [provider.domain, ...(provider.altDomains || [])];
  return domains.some(
    (d) => hostname === d.replace(/^\./, "") || hostname.endsWith(d)
  );
}

async function captureCookies(provider, btn) {
  btn.disabled = true;
  btn.textContent = "Capturing...";

  const status = document.getElementById("status");
  status.textContent = "";
  status.className = "status";

  try {
    // Get all cookies for the provider's domain(s)
    const domains = [provider.domain, ...(provider.altDomains || [])];
    const allCookies = [];

    for (const domain of domains) {
      const cookies = await chrome.cookies.getAll({ domain });
      allCookies.push(...cookies);
    }

    if (allCookies.length === 0) {
      throw new Error(
        `No cookies found for ${provider.name}. Make sure you're logged in.`
      );
    }

    // Convert to key-value map
    const cookieMap = {};
    for (const c of allCookies) {
      cookieMap[c.name] = c.value;
    }

    // Send to portal
    const portalUrl = getPortalUrl();
    const res = await fetch(`${portalUrl}/api/integrations/save-cookies`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        provider: provider.id,
        cookies: cookieMap,
        label: `${provider.name} Account`,
      }),
    });

    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      throw new Error(data.error || `Portal returned ${res.status}`);
    }

    btn.textContent = "Captured!";
    btn.classList.add("success");
    status.textContent = `${provider.name}: ${allCookies.length} cookies saved`;
    status.className = "status success";

    setTimeout(() => {
      btn.textContent = "Capture";
      btn.classList.remove("success");
      btn.disabled = false;
    }, 3000);
  } catch (err) {
    btn.textContent = "Failed";
    btn.classList.add("error");
    status.textContent = err.message;
    status.className = "status error";

    setTimeout(() => {
      btn.textContent = "Capture";
      btn.classList.remove("error");
      btn.disabled = false;
    }, 3000);
  }
}

async function render() {
  const container = document.getElementById("providers");
  const currentDomain = await getCurrentTabDomain();

  for (const provider of PROVIDERS) {
    const isActive = currentDomain && matchesDomain(currentDomain, provider);

    const div = document.createElement("div");
    div.className = `provider${isActive ? " detected" : ""}`;

    const info = document.createElement("div");
    const name = document.createElement("div");
    name.className = "provider-name";
    name.textContent = provider.name;
    const domain = document.createElement("div");
    domain.className = "provider-domain";
    domain.textContent = provider.domain;
    info.appendChild(name);
    info.appendChild(domain);

    const btn = document.createElement("button");
    btn.className = "capture-btn";
    btn.textContent = "Capture";
    btn.addEventListener("click", () => captureCookies(provider, btn));

    div.appendChild(info);
    div.appendChild(btn);
    container.appendChild(div);
  }

  // Load saved portal URL
  chrome.storage?.local?.get("portalUrl", (data) => {
    if (data?.portalUrl) {
      document.getElementById("portalUrl").value = data.portalUrl;
    }
  });

  // Save portal URL on change
  document.getElementById("portalUrl").addEventListener("change", (e) => {
    chrome.storage?.local?.set({ portalUrl: e.target.value });
  });
}

render();
