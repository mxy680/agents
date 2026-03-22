/**
 * Emdash Integrations — Popup Script
 *
 * Shows sync status for each provider. Configuration (portal URL + token)
 * is handled automatically by the portal webpage via externally_connectable,
 * so users never need to paste anything manually.
 */

"use strict"

const PROVIDER_META = {
  instagram: { label: "Instagram", icon: "📸" },
  linkedin: { label: "LinkedIn", icon: "💼" },
  x: { label: "X", icon: "𝕏" },
  canvas: { label: "Canvas LMS", icon: "🎓" },
}

// ---------------------------------------------------------------------------
// Status rendering
// ---------------------------------------------------------------------------

function formatTime(isoString) {
  const diff = Date.now() - new Date(isoString).getTime()
  if (diff < 60_000) return "just now"
  if (diff < 3_600_000) return `${Math.floor(diff / 60_000)}m ago`
  if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)}h ago`
  return new Date(isoString).toLocaleDateString()
}

function renderStatus(providerKey, statusData) {
  const el = document.getElementById(`status-${providerKey}`)
  if (!el) return

  if (!statusData) {
    el.className = "provider-status status-muted"
    el.textContent = "Not synced yet"
    return
  }

  const { status, message, syncedAt } = statusData

  if (status === "synced" && syncedAt) {
    el.className = "provider-status status-synced"
    el.textContent = `Synced ${formatTime(syncedAt)}`
  } else if (status === "not_logged_in") {
    el.className = "provider-status status-warning"
    el.textContent = "Not logged in"
  } else if (status === "error") {
    el.className = "provider-status status-error"
    el.textContent = message ?? "Error"
  } else {
    el.className = "provider-status status-muted"
    el.textContent = "Not synced yet"
  }
}

// ---------------------------------------------------------------------------
// Provider cards
// ---------------------------------------------------------------------------

function renderProviders(syncStatus) {
  const container = document.getElementById("content")
  container.innerHTML = ""

  for (const [key, meta] of Object.entries(PROVIDER_META)) {
    const card = document.createElement("div")
    card.className = "provider-card"
    card.innerHTML = `
      <div class="provider-left">
        <span class="provider-icon">${meta.icon}</span>
        <div>
          <div class="provider-name">${meta.label}</div>
          <div class="provider-status status-muted" id="status-${key}">Not synced yet</div>
        </div>
      </div>
      <button class="sync-btn" id="sync-btn-${key}">Sync</button>
    `
    container.appendChild(card)
    renderStatus(key, syncStatus?.[key])

    document.getElementById(`sync-btn-${key}`).addEventListener("click", () => {
      const btn = document.getElementById(`sync-btn-${key}`)
      btn.disabled = true
      btn.textContent = "..."
      chrome.runtime.sendMessage({ type: "sync", provider: key }, () => {
        btn.disabled = false
        btn.textContent = "Sync"
      })
    })
  }

  // Portal link
  chrome.storage.local.get("portalUrl", ({ portalUrl }) => {
    if (portalUrl) {
      const link = document.createElement("a")
      link.className = "portal-link"
      link.textContent = "Open Portal"
      link.addEventListener("click", (e) => {
        e.preventDefault()
        chrome.tabs.create({ url: `${portalUrl}/integrations` })
      })
      container.appendChild(link)
    }
  })
}

function renderNotConfigured() {
  const container = document.getElementById("content")
  container.innerHTML = `
    <div class="not-configured">
      <strong>Not connected to a portal yet</strong>
      Click "Connect" on any integration in your Emdash portal to set up automatically.
    </div>
  `
}

// ---------------------------------------------------------------------------
// Initialization
// ---------------------------------------------------------------------------

async function init() {
  const { portalUrl, token, syncStatus } = await chrome.storage.local.get([
    "portalUrl",
    "token",
    "syncStatus",
  ])

  const statusEl = document.getElementById("connection-status")

  if (!portalUrl || !token) {
    statusEl.textContent = "Not connected to a portal"
    renderNotConfigured()
  } else {
    // Show truncated portal URL
    try {
      const url = new URL(portalUrl)
      statusEl.textContent = `Connected to ${url.hostname}`
    } catch {
      statusEl.textContent = `Connected to ${portalUrl}`
    }
    renderProviders(syncStatus)
  }

  // Live updates
  chrome.storage.onChanged.addListener((changes, area) => {
    if (area !== "local") return

    if (changes.syncStatus) {
      const newStatus = changes.syncStatus.newValue ?? {}
      for (const key of Object.keys(PROVIDER_META)) {
        renderStatus(key, newStatus[key])
      }
    }

    // If portal URL or token changes, re-render everything
    if (changes.portalUrl || changes.token) {
      init()
    }
  })
}

document.addEventListener("DOMContentLoaded", init)
