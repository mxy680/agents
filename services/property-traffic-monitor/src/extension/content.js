/**
 * Content script — runs on property research sites.
 * Extracts address/property info from the page and reports to the background worker.
 */

(function () {
  "use strict";

  const SITE_EXTRACTORS = {
    "www.propertyshark.com": extractPropertyShark,
    "zola.planning.nyc.gov": extractZola,
    "a810-bisweb.nyc.gov": extractDOB,
    "a836-acris.nyc.gov": extractACRIS,
    "propertyinformationportal.nyc.gov": extractNYCPortal,
    "www.actovia.com": extractActovia,
    "www.costar.com": extractCoStar,
  };

  function extractPropertyShark() {
    // URL pattern: /mason/Property/350-5th-Ave-New-York-NY-10118/
    const match = window.location.pathname.match(/\/Property\/([^/]+)/i);
    if (match) {
      return match[1].replace(/-/g, " ");
    }
    // Fallback: page title
    const title = document.querySelector("h1");
    return title ? title.textContent.trim() : null;
  }

  function extractZola() {
    // ZoLa stores BBL in the URL hash or address bar
    const hash = window.location.hash;
    const bblMatch = hash.match(/bbl=(\d+)/);
    if (bblMatch) return `BBL:${bblMatch[1]}`;
    // Try address from search input
    const input = document.querySelector('input[type="text"]');
    return input ? input.value : null;
  }

  function extractDOB() {
    // BIS pages show address in the page content
    const addrEl = document.querySelector("td.maininfo");
    if (addrEl) return addrEl.textContent.trim();
    // Fallback: URL params
    const params = new URLSearchParams(window.location.search);
    const houseno = params.get("houseno");
    const street = params.get("street");
    if (houseno && street) return `${houseno} ${street}`;
    return null;
  }

  function extractACRIS() {
    // ACRIS shows BBL or address in the page
    const params = new URLSearchParams(window.location.search);
    const bbl = params.get("bbl");
    if (bbl) return `BBL:${bbl}`;
    return document.title || null;
  }

  function extractNYCPortal() {
    const h1 = document.querySelector("h1");
    return h1 ? h1.textContent.trim() : null;
  }

  function extractActovia() {
    const h1 = document.querySelector("h1");
    return h1 ? h1.textContent.trim() : null;
  }

  function extractCoStar() {
    const h1 = document.querySelector("h1");
    return h1 ? h1.textContent.trim() : null;
  }

  // Main
  const hostname = window.location.hostname;
  const extractor = SITE_EXTRACTORS[hostname];

  if (extractor) {
    const address = extractor();
    if (address) {
      chrome.runtime.sendMessage({
        type: "PROPERTY_VISIT",
        data: {
          address: address,
          site: hostname,
          url: window.location.href,
          timestamp: new Date().toISOString(),
        },
      });
    }
  }
})();
