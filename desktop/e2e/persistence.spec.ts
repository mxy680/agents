import { test, expect, _electron, type ElectronApplication, type Page } from '@playwright/test'
import { join } from 'path'
import { execSync } from 'child_process'

const mainPath = join(__dirname, '..', 'out', 'main', 'index.js')

test.beforeAll(() => {
  execSync('pnpm run build', { cwd: join(__dirname, '..'), stdio: 'pipe' })
})

async function launchApp(): Promise<{ app: ElectronApplication; page: Page }> {
  const app = await _electron.launch({
    args: [mainPath],
    env: { ...process.env }
  })

  // Set app name so userData resolves to ~/Library/Application Support/ade
  await app.evaluate(async ({ app }) => {
    app.setName('ade')
  })

  const page = await app.firstWindow()
  await page.waitForLoadState('domcontentloaded')
  await page.waitForTimeout(2000)
  return { app, page }
}

test.describe('Persistence', () => {
  test('messages persist in postgres across app restarts', async () => {
    // Ensure postgres is running
    try {
      execSync('docker start ade-postgres 2>/dev/null || true', { stdio: 'pipe' })
      for (let i = 0; i < 10; i++) {
        try {
          execSync("PGPASSWORD=ade psql -h localhost -p 5434 -U ade -d ade -c 'SELECT 1'", { stdio: 'pipe' })
          break
        } catch { execSync('sleep 1') }
      }
    } catch {
      test.skip()
      return
    }

    // --- Session 1: send a message ---
    const { app: app1, page: page1 } = await launchApp()

    // Check if projects exist
    const projectButtons = page1.locator('aside .space-y-0\\.5 button')
    const count = await projectButtons.count()
    if (count === 0) {
      await app1.close()
      test.skip()
      return
    }

    // Open first project
    await projectButtons.first().click()
    await page1.waitForTimeout(2000)

    // Click orchestrator
    const coordBtn = page1.locator('aside button', { hasText: 'Coordinator' })
    if (await coordBtn.count() > 0) {
      await coordBtn.click()
      await page1.waitForTimeout(500)
    }

    // Send message
    const input = page1.locator('textarea')
    await input.fill('PERSISTENCE_TEST_123: just say ok')
    await input.press('Enter')

    // Wait for save
    await page1.waitForTimeout(8000)

    // Verify in postgres
    let msgCount: string
    try {
      msgCount = execSync(
        "PGPASSWORD=ade psql -h localhost -p 5434 -U ade -d ade -t -c \"SELECT count(*) FROM messages WHERE content LIKE '%PERSISTENCE_TEST_123%'\"",
        { encoding: 'utf-8' }
      ).trim()
    } catch {
      msgCount = '0'
    }

    console.log('Messages in DB after send:', msgCount)
    expect(parseInt(msgCount)).toBeGreaterThanOrEqual(1)

    await page1.screenshot({ path: 'e2e/screenshots/persistence-before-restart.png' })
    await app1.close()
    await new Promise((r) => setTimeout(r, 2000))

    // --- Session 2: verify message persisted ---
    const { app: app2, page: page2 } = await launchApp()

    // Open same project
    const projectButtons2 = page2.locator('aside .space-y-0\\.5 button')
    await projectButtons2.first().click()
    await page2.waitForTimeout(2000)

    const coordBtn2 = page2.locator('aside button', { hasText: 'Coordinator' })
    if (await coordBtn2.count() > 0) {
      await coordBtn2.click()
      await page2.waitForTimeout(1000)
    }

    // Check message is visible
    const messageVisible = await page2.locator('text=PERSISTENCE_TEST_123').isVisible({ timeout: 5000 }).catch(() => false)
    console.log('Message visible after restart:', messageVisible)
    expect(messageVisible).toBe(true)

    await page2.screenshot({ path: 'e2e/screenshots/persistence-after-restart.png' })

    // Cleanup
    try {
      execSync(
        "PGPASSWORD=ade psql -h localhost -p 5434 -U ade -d ade -c \"DELETE FROM messages WHERE content LIKE '%PERSISTENCE_TEST_123%'\"",
        { stdio: 'pipe' }
      )
    } catch { /* ok */ }

    await app2.close()
  })
})
