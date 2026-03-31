import { test, expect, _electron } from '@playwright/test'
import { join } from 'path'
import { execSync } from 'child_process'

const mainPath = join(__dirname, '..', 'out', 'main', 'index.js')

test.beforeAll(() => {
  execSync('pnpm run build', { cwd: join(__dirname, '..'), stdio: 'pipe' })
})

test('orchestrator can use integrations CLI without crashing', async () => {
  const app = await _electron.launch({
    args: [mainPath],
    env: { ...process.env }
  })

  const page = await app.firstWindow()
  await page.waitForLoadState('domcontentloaded')
  await page.waitForTimeout(2000)

  // Open first project
  const projectButtons = page.locator('aside .space-y-0\\.5 button')
  const count = await projectButtons.count()
  if (count === 0) { await app.close(); test.skip(); return }

  await projectButtons.first().click()
  await page.waitForTimeout(2000)

  // Ensure on orchestrator
  const coordBtn = page.locator('aside button', { hasText: 'Coordinator' })
  if (await coordBtn.count() > 0) {
    await coordBtn.click()
    await page.waitForTimeout(500)
  }

  // Ask to list emails — this triggers the integrations CLI
  const input = page.locator('textarea')
  await input.fill('List my 2 most recent emails using the integrations CLI. Be brief.')
  await input.press('Enter')

  // Wait for response — orchestrator needs time to call CLI
  await page.waitForTimeout(15000)

  // App should still be alive
  const isVisible = await page.locator('textarea').isVisible()
  expect(isVisible).toBe(true)

  // Check no crash happened
  const bodyText = await page.locator('body').innerText()
  expect(bodyText).not.toContain('crashed')

  await page.screenshot({ path: 'e2e/screenshots/integrations-test.png' })
  await app.close()
})
