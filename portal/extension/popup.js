const scrapeBtn = document.getElementById("scrape-btn");
const progressEl = document.getElementById("progress");
const progressBar = document.getElementById("progress-bar");
const progressZips = document.getElementById("progress-zips");
const progressListings = document.getElementById("progress-listings");
const progressStatus = document.getElementById("progress-status");
const statusEl = document.getElementById("status");
const portalUrlInput = document.getElementById("portal-url");

const NYC_ZIP_CODES = [
  ...["10451","10452","10453","10454","10455","10456","10457","10458","10459","10460",
      "10461","10462","10463","10464","10465","10466","10467","10468","10469","10470",
      "10471","10472","10473","10474","10475"].map(z => ({ borough: "Bronx", zip: z })),
  ...["11201","11203","11204","11205","11206","11207","11208","11209","11210","11211",
      "11212","11213","11214","11215","11216","11217","11218","11219","11220","11221",
      "11222","11223","11224","11225","11226","11228","11229","11230","11231","11232",
      "11233","11234","11235","11236","11237","11238","11239"].map(z => ({ borough: "Brooklyn", zip: z })),
  ...["10001","10002","10003","10009","10010","10011","10012","10013","10014","10016",
      "10019","10021","10022","10023","10024","10025","10026","10027","10028","10029",
      "10030","10031","10032","10033","10034","10035","10037","10039","10040","10128"].map(z => ({ borough: "Manhattan", zip: z })),
  ...["11101","11102","11103","11104","11105","11106","11109","11354","11355","11356",
      "11357","11358","11359","11360","11361","11362","11363","11364","11365","11366",
      "11367","11368","11369","11370","11372","11373","11374","11375","11377","11378",
      "11379","11385","11411","11412","11413","11414","11415","11416","11417","11418",
      "11419","11420","11421","11422","11423"].map(z => ({ borough: "Queens", zip: z })),
];

// Load portal URL
chrome.storage.local.get("portalURL", (result) => {
  portalUrlInput.value = result.portalURL || "http://localhost:3000";
});

portalUrlInput.addEventListener("change", () => {
  chrome.storage.local.set({ portalURL: portalUrlInput.value.trim().replace(/\/+$/, "") });
});

// Check if a scrape is already running
chrome.runtime.sendMessage({ action: "scrape-status" }, (status) => {
  if (status?.running) {
    scrapeBtn.disabled = true;
    scrapeBtn.textContent = "Scraping...";
    progressEl.classList.add("visible");
    startPolling();
  }
});

scrapeBtn.addEventListener("click", () => {
  scrapeBtn.disabled = true;
  scrapeBtn.textContent = "Scraping...";
  progressEl.classList.add("visible");
  progressStatus.textContent = "Starting...";
  progressBar.style.width = "0%";
  progressZips.textContent = "0";
  progressListings.textContent = "0";
  statusEl.textContent = "";
  statusEl.className = "status";

  chrome.runtime.sendMessage(
    { action: "start-scrape", zipCodes: NYC_ZIP_CODES },
    (response) => {
      if (response?.ok) {
        progressStatus.textContent = "Done!";
        progressBar.style.width = "100%";
        progressListings.textContent = String(response.count || 0);
        statusEl.textContent = `${response.count} listings saved`;
        statusEl.className = "status success";
      } else {
        progressStatus.textContent = "Failed";
        statusEl.textContent = response?.error || "Unknown error";
        statusEl.className = "status error";
      }
      scrapeBtn.disabled = false;
      scrapeBtn.textContent = "Scrape Zillow";
    }
  );

  startPolling();
});

function startPolling() {
  const interval = setInterval(() => {
    chrome.runtime.sendMessage({ action: "scrape-status" }, (s) => {
      if (!s?.running) { clearInterval(interval); return; }
      const pct = Math.round((s.processed / s.total) * 100);
      progressBar.style.width = `${pct}%`;
      progressZips.textContent = String(s.processed);
      progressListings.textContent = String(s.results);
      progressStatus.textContent = `${s.currentBorough} ${s.currentZip}`;
    });
  }, 1000);
}
