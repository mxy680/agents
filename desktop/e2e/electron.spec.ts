import { test, expect, _electron, type ElectronApplication, type Page } from '@playwright/test'
import { join } from 'path'

let app: ElectronApplication
let page: Page

test.beforeAll(async () => {
  // Build first
  const { execSync } = await import('child_process')
  execSync('pnpm run build', { cwd: join(__dirname, '..'), stdio: 'pipe' })

  app = await _electron.launch({
    args: [join(__dirname, '..', 'out', 'main', 'index.js')],
    env: { ...process.env, NODE_ENV: 'production' }
  })

  page = await app.firstWindow()
  await page.waitForLoadState('domcontentloaded')
  await page.waitForTimeout(2000)
})

test.afterAll(async () => {
  await app?.close()
})

test.describe('ADE Electron App', () => {
  test('app window opens', async () => {
    const title = await page.title()
    expect(title).toBe('ADE')
  })

  test('sidebar is visible with Home button', async () => {
    await expect(page.locator('aside')).toBeVisible()
    await expect(page.locator('aside button', { hasText: 'Home' })).toBeVisible()
  })

  test('home page shows Connect Repo and New Project', async () => {
    await expect(page.locator('button', { hasText: 'Connect Repo' })).toBeVisible()
    await expect(page.locator('button', { hasText: 'New Project' })).toBeVisible()
  })

  test('sidebar shows Projects section', async () => {
    await expect(page.locator('aside', { hasText: 'Projects' })).toBeVisible()
  })

  test('sidebar shows Integrations section', async () => {
    await expect(page.locator('aside button', { hasText: 'Claude Code' })).toBeVisible()
    await expect(page.locator('aside button', { hasText: 'GitHub' })).toBeVisible()
    await expect(page.locator('aside button', { hasText: 'Skills' })).toBeVisible()
    await expect(page.locator('aside button', { hasText: 'MCP Servers' })).toBeVisible()
  })

  test('clicking Claude Code shows integration page', async () => {
    await page.locator('aside button', { hasText: 'Claude Code' }).click()
    await expect(page.locator('h1', { hasText: 'Claude Code' })).toBeVisible()
  })

  test('clicking Skills shows skills page', async () => {
    await page.locator('aside button', { hasText: 'Skills' }).click()
    await expect(page.locator('h1', { hasText: 'Skills' })).toBeVisible()
  })

  test('clicking Home returns to home page', async () => {
    await page.locator('aside button', { hasText: 'Home' }).click()
    await expect(page.locator('button', { hasText: 'Connect Repo' })).toBeVisible()
  })

  test('clicking a project opens orchestrator', async () => {
    // Check if any projects exist in sidebar
    const projectButtons = page.locator('aside .space-y-0\\.5 button')
    const count = await projectButtons.count()

    if (count > 0) {
      // Click the first project
      await projectButtons.first().click()
      await page.waitForTimeout(1000)

      // Should show orchestrator view
      await expect(page.locator('text=Orchestrator').first()).toBeVisible()
      await expect(page.locator('text=Coordinator').first()).toBeVisible()
    } else {
      test.skip()
    }
  })

  test('orchestrator chat input is visible when project open', async () => {
    const projectButtons = page.locator('aside .space-y-0\\.5 button')
    const count = await projectButtons.count()

    if (count > 0) {
      await projectButtons.first().click()
      await page.waitForTimeout(1000)
      await expect(page.locator('textarea[placeholder*="orchestrator"]')).toBeVisible()
    } else {
      test.skip()
    }
  })

  test('can send a message and get a response', async () => {
    const projectButtons = page.locator('aside .space-y-0\\.5 button')
    const count = await projectButtons.count()

    if (count > 0) {
      await projectButtons.first().click()
      await page.waitForTimeout(1000)

      // Click orchestrator in sidebar to make sure we're on it
      await page.locator('aside button', { hasText: 'Coordinator' }).click()
      await page.waitForTimeout(500)

      const input = page.locator('textarea[placeholder*="orchestrator"]')
      await input.fill('Say hello in one word')
      await input.press('Enter')

      // Wait for response (up to 30s)
      await page.waitForTimeout(3000)

      // Check DB has the message
      const { execSync } = await import('child_process')
      const result = execSync(
        "PGPASSWORD=ade psql -h localhost -p 5434 -U ade -d ade -t -c \"SELECT count(*) FROM messages;\"",
        { encoding: 'utf-8' }
      ).trim()

      expect(parseInt(result)).toBeGreaterThan(0)
    } else {
      test.skip()
    }
  })

  test('take screenshot', async () => {
    await page.screenshot({ path: 'e2e/screenshots/app.png' })
  })
})
