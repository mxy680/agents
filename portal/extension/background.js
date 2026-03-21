/**
 * Emdash Integrations — Background Service Worker
 *
 * Watches for cookie changes on target provider domains and automatically
 * syncs the session cookies to the Emdash portal via the extension API.
 */

"use strict"

// ---------------------------------------------------------------------------
// Provider definitions
// ---------------------------------------------------------------------------

const PROVIDER_LOGIN_URLS = {
  instagram: "https://www.instagram.com/accounts/login/",
  linkedin: "https://www.linkedin.com/login",
  x: "https://x.com/i/flow/login",
}

const PROVIDERS = {
  instagram: {
    domain: ".instagram.com",
    loginCookie: "sessionid",
    cookies: ["sessionid", "csrftoken", "ds_user_id", "mid", "ig_did"],
    credentialKeys: {
      sessionid: "session_id",
      csrftoken: "csrf_token",
      ds_user_id: "ds_user_id",
      mid: "mid",
      ig_did: "ig_did",
    },
  },
  linkedin: {
    domain: ".linkedin.com",
    loginCookie: "li_at",
    cookies: ["li_at", "JSESSIONID", "bcookie", "lidc", "li_mc"],
    credentialKeys: {
      li_at: "li_at",
      JSESSIONID: "jsessionid",
      bcookie: "bcookie",
      lidc: "lidc",
      li_mc: "li_mc",
    },
  },
  x: {
    domain: ".x.com",
    loginCookie: "auth_token",
    cookies: ["auth_token", "ct0"],
    credentialKeys: {
      auth_token: "auth_token",
      ct0: "csrf_token",
    },
  },
}

// Debounce timers per provider — avoids firing multiple syncs during a single
// login flow that sets several cookies in rapid succession.
const DEBOUNCE_MS = 500
const debounceTimers = {}

// ---------------------------------------------------------------------------
// Cookie domain matching
// ---------------------------------------------------------------------------

/**
 * Returns the provider key whose domain matches the given cookie domain, or
 * null if no provider matches.
 *
 * Chrome sets cookie.domain to e.g. ".instagram.com" or "www.instagram.com".
 * We match by checking whether the cookie domain ends with the provider domain
 * (stripping the leading dot for a suffix check).
 */
function findProviderForCookie(cookieDomain) {
  for (const [key, provider] of Object.entries(PROVIDERS)) {
    const suffix = provider.domain.startsWith(".")
      ? provider.domain.slice(1)
      : provider.domain
    if (cookieDomain === provider.domain || cookieDomain.endsWith(suffix)) {
      return key
    }
  }
  return null
}

// ---------------------------------------------------------------------------
// Sync logic
// ---------------------------------------------------------------------------

/**
 * Collects all relevant cookies for a provider and posts them to the portal.
 */
async function syncProvider(providerKey) {
  const provider = PROVIDERS[providerKey]

  // Load settings
  const { portalUrl, token } = await chrome.storage.local.get(["portalUrl", "token"])
  if (!portalUrl || !token) {
    await setStatus(providerKey, {
      status: "error",
      message: "Portal URL or token not configured",
      syncedAt: null,
    })
    return
  }

  // Gather cookies using URL-based filter for precision
  const url = `https://${provider.domain.replace(/^\./, "www.")}`
  let allCookies
  try {
    allCookies = await chrome.cookies.getAll({ url })
  } catch (err) {
    await setStatus(providerKey, {
      status: "error",
      message: `Failed to read cookies: ${err.message}`,
      syncedAt: null,
    })
    return
  }

  // Build a name → value map for the cookies we care about
  const cookieMap = {}
  for (const cookie of allCookies) {
    if (provider.cookies.includes(cookie.name)) {
      cookieMap[cookie.name] = cookie.value
    }
  }

  // Check login cookie exists
  if (!cookieMap[provider.loginCookie]) {
    await setStatus(providerKey, {
      status: "not_logged_in",
      message: "Not logged in",
      syncedAt: null,
    })
    return
  }

  // Map raw cookie names to credential keys expected by the portal
  const credentials = {}
  for (const [rawName, credKey] of Object.entries(provider.credentialKeys)) {
    if (cookieMap[rawName]) {
      credentials[credKey] = cookieMap[rawName]
    }
  }

  // POST to portal
  try {
    const response = await fetch(`${portalUrl}/api/integrations/extension/cookies`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({ provider: providerKey, cookies: credentials }),
    })

    if (!response.ok) {
      const data = await response.json().catch(() => ({}))
      const message = data.error ?? `HTTP ${response.status}`
      await setStatus(providerKey, { status: "error", message, syncedAt: null })
      return
    }

    await setStatus(providerKey, {
      status: "synced",
      message: null,
      syncedAt: new Date().toISOString(),
    })
  } catch (err) {
    await setStatus(providerKey, {
      status: "error",
      message: `Network error: ${err.message}`,
      syncedAt: null,
    })
  }
}

/**
 * Persists sync status for a provider to chrome.storage.local so the popup
 * can read it.
 */
async function setStatus(providerKey, { status, message, syncedAt }) {
  const { syncStatus = {} } = await chrome.storage.local.get("syncStatus")
  syncStatus[providerKey] = { status, message, syncedAt }
  await chrome.storage.local.set({ syncStatus })
}

/**
 * Schedules a debounced sync for the given provider.
 */
function scheduleSyncProvider(providerKey) {
  if (debounceTimers[providerKey]) {
    clearTimeout(debounceTimers[providerKey])
  }
  debounceTimers[providerKey] = setTimeout(async () => {
    delete debounceTimers[providerKey]
    await syncProvider(providerKey)
  }, DEBOUNCE_MS)
}

// ---------------------------------------------------------------------------
// Event listeners
// ---------------------------------------------------------------------------

// On install/update, reload any open portal tabs so the content script injects
// and the user doesn't need to manually refresh.
chrome.runtime.onInstalled.addListener(() => {
  chrome.tabs.query({}, (tabs) => {
    for (const tab of tabs) {
      if (!tab.url) continue
      // Match the same origins as our content_scripts.matches
      if (
        tab.url.startsWith("http://localhost") ||
        tab.url.includes(".emdash.io") ||
        tab.url.includes(".emdash.dev")
      ) {
        chrome.tabs.reload(tab.id)
      }
    }
  })
})

// Watch for cookie changes on target domains
chrome.cookies.onChanged.addListener((changeInfo) => {
  const { cookie, removed } = changeInfo
  if (removed) return // Only care about cookies being set

  const providerKey = findProviderForCookie(cookie.domain)
  if (!providerKey) return

  const provider = PROVIDERS[providerKey]

  // Only trigger on a tracked cookie name for this provider
  if (!provider.cookies.includes(cookie.name)) return

  scheduleSyncProvider(providerKey)

  // Also check if this cookie change matches a pending incognito login
  chrome.storage.session.get("pendingLogins", async ({ pendingLogins = {} }) => {
    for (const [sessionId, session] of Object.entries(pendingLogins)) {
      if (session.provider === providerKey && session.status === "waiting") {
        // Login cookie detected for a pending session — capture!
        if (cookie.name === provider.loginCookie) {
          session.status = "capturing"
          await chrome.storage.session.set({ pendingLogins })

          // Debounce the capture (cookies come in bursts during login)
          setTimeout(() => captureIncognitoCookies(sessionId), 1000)
        }
        break
      }
    }
  })
})

// Handle messages from the popup
chrome.runtime.onMessage.addListener((message, _sender, sendResponse) => {
  if (message.type === "sync" && message.provider) {
    syncProvider(message.provider)
      .then(() => sendResponse({ ok: true }))
      .catch((err) => sendResponse({ ok: false, error: err.message }))
    return true // keep channel open for async response
  }

  if (message.type === "sync-all") {
    Promise.all(Object.keys(PROVIDERS).map((key) => syncProvider(key)))
      .then(() => sendResponse({ ok: true }))
      .catch((err) => sendResponse({ ok: false, error: err.message }))
    return true
  }
})

// ---------------------------------------------------------------------------
// External messages from the portal webpage (via externally_connectable)
// ---------------------------------------------------------------------------

chrome.runtime.onMessageExternal.addListener((message, sender, sendResponse) => {
  // "configure" — portal sends token + URL so user doesn't have to copy/paste
  if (message.type === "configure" && message.token && message.portalUrl) {
    chrome.storage.local.set(
      { portalUrl: message.portalUrl, token: message.token },
      () => sendResponse({ ok: true })
    )
    return true
  }

  // "sync" — portal asks extension to sync a specific provider right now
  if (message.type === "sync" && message.provider && PROVIDERS[message.provider]) {
    syncProvider(message.provider)
      .then(() => sendResponse({ ok: true, status: "synced" }))
      .catch((err) => sendResponse({ ok: false, error: err.message }))
    return true
  }

  // "ping" — portal checks if extension is installed and reachable
  if (message.type === "ping") {
    sendResponse({ ok: true, version: chrome.runtime.getManifest().version })
    return false
  }

  // "status" — portal asks for current sync status
  if (message.type === "status") {
    chrome.storage.local.get("syncStatus", ({ syncStatus }) => {
      sendResponse({ ok: true, syncStatus: syncStatus ?? {} })
    })
    return true
  }

  // "check-incognito" — check if extension has incognito access
  if (message.type === "check-incognito") {
    chrome.extension.isAllowedIncognitoAccess((allowed) => {
      sendResponse({ ok: true, allowed })
    })
    return true
  }

  // "login" — open incognito window for login
  if (message.type === "login" && message.provider && PROVIDERS[message.provider]) {
    const sessionId = Math.random().toString(36).slice(2) + Date.now().toString(36)
    const loginUrl = PROVIDER_LOGIN_URLS[message.provider]

    chrome.extension.isAllowedIncognitoAccess(async (allowed) => {
      if (!allowed) {
        sendResponse({ ok: false, error: "Incognito access not enabled" })
        return
      }

      try {
        const win = await chrome.windows.create({ incognito: true, url: loginUrl })

        const session = {
          sessionId,
          provider: message.provider,
          label: message.label || `${message.provider} Account`,
          windowId: win.id,
          status: "waiting",
          error: null,
          createdAt: Date.now(),
        }

        const { pendingLogins = {} } = await chrome.storage.session.get("pendingLogins")
        pendingLogins[sessionId] = session
        await chrome.storage.session.set({ pendingLogins })

        sendResponse({ ok: true, sessionId })
      } catch (err) {
        sendResponse({ ok: false, error: err.message })
      }
    })
    return true
  }

  // "login-status" — check status of a login session
  if (message.type === "login-status" && message.sessionId) {
    chrome.storage.session.get("pendingLogins", ({ pendingLogins = {} }) => {
      const session = pendingLogins[message.sessionId]
      if (!session) {
        sendResponse({ ok: false, error: "Session not found" })
      } else {
        sendResponse({ ok: true, status: session.status, error: session.error })
      }
    })
    return true
  }

  // "cancel-login" — cancel and cleanup
  if (message.type === "cancel-login" && message.sessionId) {
    chrome.storage.session.get("pendingLogins", async ({ pendingLogins = {} }) => {
      const session = pendingLogins[message.sessionId]
      if (session?.windowId) {
        try { await chrome.windows.remove(session.windowId) } catch {}
      }
      delete pendingLogins[message.sessionId]
      await chrome.storage.session.set({ pendingLogins })
      sendResponse({ ok: true })
    })
    return true
  }
})

// ---------------------------------------------------------------------------
// Incognito login helpers
// ---------------------------------------------------------------------------

/**
 * Captures cookies from the incognito session and posts them to the portal.
 */
async function captureIncognitoCookies(sessionId) {
  const { pendingLogins = {} } = await chrome.storage.session.get("pendingLogins")
  const session = pendingLogins[sessionId]
  if (!session || session.status === "complete") return

  const provider = PROVIDERS[session.provider]
  const { portalUrl, token } = await chrome.storage.local.get(["portalUrl", "token"])

  if (!portalUrl || !token) {
    session.status = "error"
    session.error = "Portal not configured"
    await chrome.storage.session.set({ pendingLogins })
    return
  }

  // Get cookies — in the incognito worker, getAll returns incognito cookies by default
  const url = `https://${provider.domain.replace(/^\./, "www.")}`
  let allCookies
  try {
    allCookies = await chrome.cookies.getAll({ url })
  } catch (err) {
    session.status = "error"
    session.error = `Failed to read cookies: ${err.message}`
    await chrome.storage.session.set({ pendingLogins })
    return
  }

  // Build credential map
  const cookieMap = {}
  for (const c of allCookies) {
    if (provider.cookies.includes(c.name)) {
      cookieMap[c.name] = c.value
    }
  }

  if (!cookieMap[provider.loginCookie]) {
    session.status = "error"
    session.error = "Login cookie not found"
    await chrome.storage.session.set({ pendingLogins })
    return
  }

  const credentials = {}
  for (const [rawName, credKey] of Object.entries(provider.credentialKeys)) {
    if (cookieMap[rawName]) {
      credentials[credKey] = cookieMap[rawName]
    }
  }

  // POST to portal
  try {
    const response = await fetch(`${portalUrl}/api/integrations/extension/cookies`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token}`,
      },
      body: JSON.stringify({
        provider: session.provider,
        cookies: credentials,
        label: session.label,
      }),
    })

    if (!response.ok) {
      const data = await response.json().catch(() => ({}))
      session.status = "error"
      session.error = data.error || `HTTP ${response.status}`
    } else {
      session.status = "complete"
      session.error = null

      // Close the incognito window
      if (session.windowId) {
        try { await chrome.windows.remove(session.windowId) } catch {}
      }
    }
  } catch (err) {
    session.status = "error"
    session.error = `Network error: ${err.message}`
  }

  await chrome.storage.session.set({ pendingLogins })
}

// Detect when user closes the incognito window manually
chrome.windows.onRemoved.addListener((windowId) => {
  chrome.storage.session.get("pendingLogins", async ({ pendingLogins = {} }) => {
    for (const [, session] of Object.entries(pendingLogins)) {
      if (session.windowId === windowId && session.status === "waiting") {
        session.status = "cancelled"
        session.error = "Window closed"
        await chrome.storage.session.set({ pendingLogins })
        break
      }
    }
  })
})
