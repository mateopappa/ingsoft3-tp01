import { expect, test } from '@playwright/test'

async function registerFreshUser(page) {
  const email = `e2e-${Date.now()}-${Math.floor(Math.random() * 1e6)}@flow.test`
  const password = 'password123'
  await page.goto('/')
  await page.locator('#tab-register-btn').click()
  await page.getByLabel('Email').fill(email)
  await page.getByLabel('Password').fill(password)
  await page.locator('#auth-submit-btn').click()
  await expect(page.getByRole('button', { name: 'Log Out' })).toBeVisible()
  await expect(page.getByLabel('Timer category').locator('option', { hasText: 'Work' })).toHaveCount(1)
}

async function createActivityViaUI(page, { description, hours, minutes }) {
  await page.getByRole('button', { name: '+ Log Activity Manually' }).click()
  const modal = page.locator('#activity-modal')
  const category = modal.getByLabel('Category *')
  await expect(modal.getByLabel('Description *')).toBeVisible()
  await modal.getByLabel('Description *').fill(description)
  // Prefer value (UUID): selecting by label can race with a categories re-render.
  const workValue = await category.locator('option', { hasText: /^Work$/ }).getAttribute('value')
  expect(workValue).toBeTruthy()
  await category.selectOption(workValue)
  await expect(category).toHaveValue(workValue)
  await modal.getByLabel('Hours').fill(String(hours))
  await modal.getByLabel('Minutes').fill(String(minutes))
  // Re-assert right before submit (late loadCategories used to wipe the selection).
  await category.selectOption(workValue)

  const post = page.waitForResponse(
    (res) => res.url().includes('/api/activities') && res.request().method() === 'POST',
  )
  await modal.locator('#activity-modal-form').evaluate((form) => form.requestSubmit())
  const res = await post
  expect(res.status(), await res.text()).toBe(201)
  await expect(modal).toHaveClass(/hidden/)
}

test('crear una actividad la muestra en la lista, y borrarla la saca', async ({ page }) => {
  page.on('dialog', (dialog) => dialog.accept())
  await registerFreshUser(page)

  const description = `e2e ${Date.now()}`
  await createActivityViaUI(page, { description, hours: 0, minutes: 30 })

  const row = page.getByRole('row').filter({ hasText: description })
  await expect(row).toBeVisible()
  await row.getByRole('button', { name: `Delete ${description}` }).click()
  await expect(row).toHaveCount(0)
})

test('duración cero muestra el error y no crea nada', async ({ page }) => {
  await registerFreshUser(page)

  const description = `e2e bad ${Date.now()}`
  await page.getByRole('button', { name: '+ Log Activity Manually' }).click()
  const modal = page.locator('#activity-modal')
  await modal.getByLabel('Description *').fill(description)
  await modal.getByLabel('Category *').selectOption({ label: 'Work' })
  await modal.getByLabel('Hours').fill('0')
  await modal.getByLabel('Minutes').fill('0')
  await modal.getByRole('button', { name: 'Save Activity' }).click()

  await expect(page.getByRole('alert')).toBeVisible()
  await expect(page.getByRole('alert')).toContainText(/Duration must be greater than zero/i)
  await expect(page.getByRole('row').filter({ hasText: description })).toHaveCount(0)
})

test('la actividad sigue en el historial al recargar', async ({ page }) => {
  page.on('dialog', (dialog) => dialog.accept())
  await registerFreshUser(page)

  const description = `e2e reload ${Date.now()}`
  await createActivityViaUI(page, { description, hours: 1, minutes: 0 })
  await expect(page.getByRole('row').filter({ hasText: description })).toBeVisible()

  await page.reload()
  await expect(page.getByRole('button', { name: 'Log Out' })).toBeVisible()
  await expect(page.getByRole('row').filter({ hasText: description })).toBeVisible()

  await page.getByRole('row').filter({ hasText: description }).getByRole('button', { name: `Delete ${description}` }).click()
  await expect(page.getByRole('row').filter({ hasText: description })).toHaveCount(0)
})
