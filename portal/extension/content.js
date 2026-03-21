/**
 * Emdash Integrations — Content Script
 *
 * Injected into portal pages to:
 * 1. Announce the extension's presence via a DOM attribute (shared across worlds)
 * 2. Relay messages between page and background via window.postMessage
 *    (the only mechanism that crosses the content script isolation boundary)
 */

"use strict"

// Announce presence via shared DOM — page reads this to detect the extension
document.documentElement.setAttribute("data-emdash-extension", chrome.runtime.id)

// Relay: page → content script → background → content script → page
// Uses window.postMessage which crosses the isolation boundary (unlike CustomEvent)
window.addEventListener("message", (event) => {
  if (event.source !== window) return
  if (event.data?.direction !== "emdash-to-extension") return

  const { id, payload } = event.data

  chrome.runtime.sendMessage(payload, (response) => {
    window.postMessage(
      {
        direction: "emdash-from-extension",
        id,
        response: response ?? null,
        error: chrome.runtime.lastError?.message ?? null,
      },
      "*"
    )
  })
})
