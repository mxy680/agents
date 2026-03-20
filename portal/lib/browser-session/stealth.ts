import type { Page } from "playwright-core"

/**
 * Applies anti-detection patches to the page before any scripts run.
 * Prevents common automation fingerprinting checks used by Instagram.
 */
export async function applyStealthScripts(page: Page): Promise<void> {
  await page.addInitScript(() => {
    // Override webdriver flag
    Object.defineProperty(navigator, "webdriver", {
      get: () => false,
      configurable: true,
    })

    // Patch plugins to look like a real browser
    Object.defineProperty(navigator, "plugins", {
      get: () => {
        const plugins = [
          { name: "Chrome PDF Plugin", filename: "internal-pdf-viewer", description: "Portable Document Format" },
          { name: "Chrome PDF Viewer", filename: "mhjfbmdgcfjbbpaeojofohoefgiehjai", description: "" },
          { name: "Native Client", filename: "internal-nacl-plugin", description: "" },
        ]
        return Object.create(PluginArray.prototype, {
          length: { value: plugins.length },
          ...Object.fromEntries(
            plugins.map((p, i) => [
              i,
              {
                value: Object.create(Plugin.prototype, {
                  name: { value: p.name },
                  filename: { value: p.filename },
                  description: { value: p.description },
                  length: { value: 0 },
                }),
              },
            ])
          ),
        })
      },
      configurable: true,
    })

    // Patch languages
    Object.defineProperty(navigator, "languages", {
      get: () => ["en-US", "en"],
      configurable: true,
    })

    // Patch chrome.runtime to have an id property
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const win = window as any
    if (typeof win.chrome === "undefined") {
      Object.defineProperty(window, "chrome", {
        value: { runtime: { id: "mhjfbmdgcfjbbpaeojofohoefgiehjai" } },
        configurable: true,
      })
    } else if (!win.chrome.runtime) {
      win.chrome.runtime = { id: "mhjfbmdgcfjbbpaeojofohoefgiehjai" }
    }

    // Remove automation detection via permissions.query
    if (navigator.permissions && navigator.permissions.query) {
      const originalQuery = navigator.permissions.query.bind(navigator.permissions)
      navigator.permissions.query = (parameters: PermissionDescriptor) => {
        if (parameters.name === "notifications") {
          return Promise.resolve({ state: "denied", onchange: null } as PermissionStatus)
        }
        return originalQuery(parameters)
      }
    }

    // Spoof hardwareConcurrency (headless often reports 1-2)
    Object.defineProperty(navigator, "hardwareConcurrency", {
      get: () => 8,
      configurable: true,
    })

    // Spoof deviceMemory
    Object.defineProperty(navigator, "deviceMemory", {
      get: () => 8,
      configurable: true,
    })

    // Spoof connection type
    Object.defineProperty(navigator, "connection", {
      get: () => ({
        effectiveType: "4g",
        rtt: 50,
        downlink: 10,
        saveData: false,
      }),
      configurable: true,
    })

    // Patch WebGL renderer and vendor to match real hardware
    const getParameterOrig = WebGLRenderingContext.prototype.getParameter
    WebGLRenderingContext.prototype.getParameter = function (param: number) {
      // UNMASKED_VENDOR_WEBGL
      if (param === 0x9245) return "Google Inc. (Apple)"
      // UNMASKED_RENDERER_WEBGL
      if (param === 0x9246) return "ANGLE (Apple, Apple M1 Pro, OpenGL 4.1)"
      return getParameterOrig.call(this, param)
    }

    // Do the same for WebGL2
    if (typeof WebGL2RenderingContext !== "undefined") {
      const getParam2Orig = WebGL2RenderingContext.prototype.getParameter
      WebGL2RenderingContext.prototype.getParameter = function (param: number) {
        if (param === 0x9245) return "Google Inc. (Apple)"
        if (param === 0x9246) return "ANGLE (Apple, Apple M1 Pro, OpenGL 4.1)"
        return getParam2Orig.call(this, param)
      }
    }
  })
}
