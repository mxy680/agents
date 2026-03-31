import { test, expect } from '@playwright/test'

const BASE = 'http://localhost:5173'

test.describe('Sidebar & Home Page', () => {
  test('shows Projects section header in sidebar', async ({ page }) => {
    await page.goto(BASE)
    // Use the section header specifically
    const sidebar = page.locator('aside')
    await expect(sidebar.locator('.text-xs.uppercase', { hasText: 'Projects' })).toBeVisible()
  })

  test('shows ADE title and subtitle on home page', async ({ page }) => {
    await page.goto(BASE)
    await expect(page.locator('h1', { hasText: 'ADE' })).toBeVisible()
    await expect(page.locator('text=Agentic Development Environment')).toBeVisible()
  })

  test('shows Connect Repo and New Project cards', async ({ page }) => {
    await page.goto(BASE)
    await expect(page.locator('button', { hasText: 'Connect Repo' })).toBeVisible()
    await expect(page.locator('button', { hasText: 'New Project' })).toBeVisible()
  })

  test('does not show Orchestrator section when no project active', async ({ page }) => {
    await page.goto(BASE)
    await page.waitForTimeout(500)
    const sidebar = page.locator('aside nav')
    await expect(sidebar.locator('.text-xs.uppercase', { hasText: 'Orchestrator' })).not.toBeVisible()
  })

  test('does not show Minions section when no project active', async ({ page }) => {
    await page.goto(BASE)
    await page.waitForTimeout(500)
    const sidebar = page.locator('aside nav')
    await expect(sidebar.locator('.text-xs.uppercase', { hasText: 'Minions' })).not.toBeVisible()
  })

  test('sidebar and main area layout exists', async ({ page }) => {
    await page.goto(BASE)
    await expect(page.locator('aside')).toBeVisible()
    await expect(page.locator('main')).toBeVisible()
  })

  test('shows "No projects yet" when project list is empty', async ({ page }) => {
    await page.goto(BASE)
    // IPC fails in browser, so project list will be empty
    await expect(page.locator('button', { hasText: 'No projects yet' })).toBeVisible()
  })

  test('take screenshot of home page', async ({ page }) => {
    await page.goto(BASE)
    await page.waitForTimeout(500)
    await page.screenshot({ path: 'e2e/screenshots/home.png', fullPage: true })
  })
})
