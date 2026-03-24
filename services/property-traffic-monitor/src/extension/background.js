/**
 * Background service worker — receives property visit reports from content scripts
 * and forwards them to the monitoring API.
 */

const DEFAULT_API_URL = "http://localhost:8000";

// Get API URL from storage or use default
async function getApiUrl() {
  const result = await chrome.storage.local.get("apiUrl");
  return result.apiUrl || DEFAULT_API_URL;
}

// Handle messages from content scripts
chrome.runtime.onMessage.addListener((message, sender, sendResponse) => {
  if (message.type === "PROPERTY_VISIT") {
    reportVisit(message.data);
  }
  return true;
});

async function reportVisit(data) {
  const apiUrl = await getApiUrl();
  try {
    const response = await fetch(`${apiUrl}/extension/report`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    });
    if (response.ok) {
      console.log("[PTM] Reported visit:", data.address, data.site);
      // Update badge
      chrome.action.setBadgeText({ text: "!" });
      chrome.action.setBadgeBackgroundColor({ color: "#4CAF50" });
      setTimeout(() => chrome.action.setBadgeText({ text: "" }), 3000);
    }
  } catch (err) {
    console.warn("[PTM] Failed to report:", err.message);
  }
}
