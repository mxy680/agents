document.addEventListener("DOMContentLoaded", async () => {
  const siteEl = document.getElementById("site")
  const copyBtn = document.getElementById("copy")
  const statusEl = document.getElementById("status")

  // Get the active tab
  const [tab] = await chrome.tabs.query({ active: true, currentWindow: true })
  if (!tab?.url) {
    siteEl.textContent = "No active tab"
    copyBtn.disabled = true
    return
  }

  const url = new URL(tab.url)
  siteEl.textContent = url.hostname

  copyBtn.addEventListener("click", async () => {
    copyBtn.disabled = true
    copyBtn.textContent = "Copying..."
    statusEl.textContent = ""

    try {
      // Read all cookies for this site
      const cookies = await chrome.cookies.getAll({ url: tab.url })

      // Build a simple name:value map
      const cookieMap = {}
      for (const c of cookies) {
        cookieMap[c.name] = c.value
      }

      // Include the site URL so the portal knows which Canvas instance
      const payload = JSON.stringify({
        _site_url: url.origin,
        ...cookieMap,
      })

      await navigator.clipboard.writeText(payload)

      copyBtn.textContent = "Copied!"
      statusEl.className = "status success"
      statusEl.textContent = `${cookies.length} cookies copied`

      setTimeout(() => {
        copyBtn.textContent = "Copy Cookies"
        copyBtn.disabled = false
      }, 2000)
    } catch (err) {
      copyBtn.textContent = "Copy Cookies"
      copyBtn.disabled = false
      statusEl.className = "status error"
      statusEl.textContent = err.message
    }
  })
})
