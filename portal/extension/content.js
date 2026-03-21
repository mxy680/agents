/**
 * Emdash Integrations — Content Script
 *
 * Injected into portal pages to announce the extension's presence.
 * The portal page reads window.__EMDASH_EXTENSION_ID__ to know which
 * extension ID to use with chrome.runtime.sendMessage().
 *
 * Also relays messages between the portal page and the background
 * service worker, since the page context can't call chrome.runtime
 * directly but the content script can.
 */

"use strict"

// Inject the extension ID into the page context
const script = document.createElement("script")
script.textContent = `window.__EMDASH_EXTENSION_ID__ = "${chrome.runtime.id}";`
;(document.head || document.documentElement).appendChild(script)
script.remove()

// Relay messages from the page to the background service worker
window.addEventListener("message", (event) => {
  // Only accept messages from the same page
  if (event.source !== window) return
  if (!event.data || event.data.direction !== "emdash-to-extension") return

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
