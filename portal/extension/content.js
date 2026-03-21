/**
 * Emdash Integrations — Content Script
 *
 * Injected into portal pages to announce the extension's presence and relay
 * messages between the portal page and the background service worker.
 *
 * Communication uses CustomEvents on the document, which work across the
 * content script / page context boundary without needing inline script
 * injection (which can be blocked by CSP).
 */

"use strict"

// Announce presence by setting a data attribute on the document element.
// The portal page can check for this synchronously.
document.documentElement.setAttribute("data-emdash-extension", chrome.runtime.id)

// Also dispatch an event for pages that are already loaded
document.dispatchEvent(
  new CustomEvent("emdash-extension-ready", {
    detail: { id: chrome.runtime.id },
  })
)

// Relay messages from the page to the background service worker.
// The page dispatches "emdash-to-extension" events on the document,
// and we respond with "emdash-from-extension" events.
document.addEventListener("emdash-to-extension", (event) => {
  const { id, payload } = event.detail

  chrome.runtime.sendMessage(payload, (response) => {
    document.dispatchEvent(
      new CustomEvent("emdash-from-extension", {
        detail: {
          id,
          response: response ?? null,
          error: chrome.runtime.lastError?.message ?? null,
        },
      })
    )
  })
})
