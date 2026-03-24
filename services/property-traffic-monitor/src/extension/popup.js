const apiInput = document.getElementById("apiUrl");
const saveBtn = document.getElementById("save");
const statusDiv = document.getElementById("status");

// Load saved URL
chrome.storage.local.get("apiUrl", (result) => {
  apiInput.value = result.apiUrl || "http://localhost:8000";
  checkConnection(apiInput.value);
});

// Save URL
saveBtn.addEventListener("click", () => {
  const url = apiInput.value.replace(/\/+$/, "");
  chrome.storage.local.set({ apiUrl: url }, () => {
    checkConnection(url);
  });
});

async function checkConnection(url) {
  try {
    const resp = await fetch(`${url}/health`);
    if (resp.ok) {
      statusDiv.textContent = "Connected to API server";
      statusDiv.className = "status ok";
    } else {
      statusDiv.textContent = `Server returned ${resp.status}`;
      statusDiv.className = "status error";
    }
  } catch (err) {
    statusDiv.textContent = "Cannot reach API server";
    statusDiv.className = "status error";
  }
}
