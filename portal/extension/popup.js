/**
 * Emdash Integrations — Popup Script
 */

"use strict"

// ---------------------------------------------------------------------------
// Provider display metadata
// ---------------------------------------------------------------------------

const PROVIDER_META = {
  instagram: { label: "Instagram", icon: "📸" },
  linkedin: { label: "LinkedIn", icon: "💼" },
  x: { label: "X", icon: "𝕏" },
}

// ---------------------------------------------------------------------------
// Debounce helper
// ---------------------------------------------------------------------------

function debounce(fn, ms) {
  let timer
  return (...args) => {
    clearTimeout(timer)
    timer = setTimeout(() => fn(...args), ms)
  }
}

// ---------------------------------------------------------------------------
// Status rendering
// ---------------------------------------------------------------------------

function renderStatus(providerKey, statusData) {
  const el = document.getElementById(`status-${providerKey}`)
  if (!el) return

  if (!statusData) {
    el.className = "provider-status status-muted"
    el.textContent = "Never synced"
    return
  }

  const { status, message, syncedAt } = statusData

  if (status === "synced" && syncedAt) {
    const when = new Date(syncedAt)
    const diff = Date.now() - when.getTime()
    let timeStr
    if (diff < 60_000) {
      timeStr = "just now"
    } else if (diff < 3_600_000) {
      timeStr = `${Math.floor(diff / 60_000)}m ago`
    } else if (diff < 86_400_000) {
      timeStr = `${Math.floor(diff / 3_600_000)}h ago`
    } else {
      timeStr = when.toLocaleDateString()
    }
    el.className = "provider-status status-synced"
    el.textContent = `Synced ${timeStr}`
  } else if (status === "not_logged_in") {
    el.className = "provider-status status-warning"
    el.textContent = "Not logged in"
  } else if (status === "error") {
    el.className = "provider-status status-error"
    el.textContent = `Error: ${message ?? "Unknown error"}`
  } else {
    el.className = "provider-status status-muted"
    el.textContent = "Never synced"
  }
}

// ---------------------------------------------------------------------------
// Provider card rendering
// ---------------------------------------------------------------------------

function renderProviders(syncStatus) {
  const container = document.getElementById("providers")
  container.innerHTML = ""

  for (const [key, meta] of Object.entries(PROVIDER_META)) {
    const card = document.createElement("div")
    card.className = "provider-card"
    card.innerHTML = `
      <div class="provider-left">
        <span class="provider-icon">${meta.icon}</span>
        <div class="provider-info">
          <div class="provider-name">${meta.label}</div>
          <div class="provider-status status-muted" id="status-${key}">Never synced</div>
        </div>
      </div>
      <button class="sync-btn" id="sync-btn-${key}">Sync Now</button>
    `
    container.appendChild(card)

    // Render initial status
    renderStatus(key, syncStatus?.[key])

    // Sync Now button handler
    document.getElementById(`sync-btn-${key}`).addEventListener("click", () => {
      setSyncing(key, true)
      chrome.runtime.sendMessage({ type: "sync", provider: key }, () => {
        setSyncing(key, false)
        // Status will update via storage listener
      })
    })
  }
}

function setSyncing(providerKey, isSyncing) {
  const btn = document.getElementById(`sync-btn-${providerKey}`)
  if (!btn) return
  btn.disabled = isSyncing
  btn.textContent = isSyncing ? "Syncing…" : "Sync Now"
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

  // Populate inputs
  const portalInput = document.getElementById("portal-url")
  const tokenInput = document.getElementById("auth-token")
  portalInput.value = portalUrl ?? ""
  tokenInput.value = token ?? ""

  // Render provider cards
  renderProviders(syncStatus)

  // Save portal URL on change (debounced)
  portalInput.addEventListener(
    "input",
    debounce(async () => {
      await chrome.storage.local.set({ portalUrl: portalInput.value.trim() })
    }, 600)
  )

  // Save token on change (debounced)
  tokenInput.addEventListener(
    "input",
    debounce(async () => {
      await chrome.storage.local.set({ token: tokenInput.value.trim() })
    }, 600)
  )

  // Open portal link
  document.getElementById("open-portal").addEventListener("click", () => {
    const url = portalInput.value.trim()
    if (!url) return
    const dest = url.endsWith("/") ? `${url}integrations` : `${url}/integrations`
    chrome.tabs.create({ url: dest })
  })

  // Sync All button
  const syncAllBtn = document.getElementById("sync-all")
  syncAllBtn.addEventListener("click", () => {
    syncAllBtn.disabled = true
    syncAllBtn.textContent = "Syncing…"
    chrome.runtime.sendMessage({ type: "sync-all" }, () => {
      syncAllBtn.disabled = false
      syncAllBtn.textContent = "Sync All"
    })
  })

  // Listen for storage changes to update status in real-time
  chrome.storage.onChanged.addListener((changes, area) => {
    if (area !== "local") return
    if (changes.syncStatus) {
      const newStatus = changes.syncStatus.newValue ?? {}
      for (const key of Object.keys(PROVIDER_META)) {
        renderStatus(key, newStatus[key])
      }
    }
    if (changes.portalUrl) {
      portalInput.value = changes.portalUrl.newValue ?? ""
    }
    if (changes.token) {
      tokenInput.value = changes.token.newValue ?? ""
    }
  })
}

document.addEventListener("DOMContentLoaded", init)
