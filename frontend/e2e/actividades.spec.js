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
  // Wait until post-login loads settle (avoids racing create vs empty list).
  await expect(page.getByLabel('Timer category').locator('option', { hasText: 'Work' })).toHaveCount(1)
  await expect(page.locator('#history-count-badge')).toContainText(/activities/)
}

async function openManualLog(page) {
  await page.getByRole('button', { name: '+ Log Activity Manually' }).click()
  const modal = page.locator('#activity-modal')
  await expect(modal).not.toHaveClass(/hidden/)
  // Exact label: getByLabel('Description') also matches "Timer description".
  await expect(modal.getByLabel('Description *')).toBeVisible()
  await expect(modal.getByLabel('Category *').locator('option', { hasText: 'Work' })).toHaveCount(1)
}

async function fillManualActivity(page, { description, hours, minutes }) {
  const modal = page.locator('#activity-modal')
  await modal.getByLabel('Description *').fill(description)
  await modal.getByLabel('Category *').selectOption({ label: 'Work' })
  await modal.getByLabel('Hours').fill(hours)
  await modal.getByLabel('Minutes').fill(minutes)
}

async function saveManualActivity(page) {
  const modal = page.locator('#activity-modal')
  await modal.getByRole('button', { name: 'Save Activity' }).click()
  await expect(modal).toHaveClass(/hidden/)
}

test('crear una actividad la muestra en la lista, y borrarla la saca', async ({ page }) => {
  page.on('dialog', (dialog) => dialog.accept())
  await registerFreshUser(page)

  const description = `e2e ${Date.now()}`
  await openManualLog(page)
  await fillManualActivity(page, { description, hours: '0', minutes: '30' })
  await saveManualActivity(page)

  const row = page.getByRole('row').filter({ hasText: description })
  await expect(row).toBeVisible()

  await row.getByRole('button', { name: `Delete ${description}` }).click()
  await expect(row).toHaveCount(0)
})

test('duración cero muestra el error y no crea nada', async ({ page }) => {
  await registerFreshUser(page)

  const description = `e2e bad ${Date.now()}`
  await openManualLog(page)
  await fillManualActivity(page, { description, hours: '0', minutes: '0' })
  await page.locator('#activity-modal').getByRole('button', { name: 'Save Activity' }).click()

  await expect(page.getByRole('alert')).toBeVisible()
  await expect(page.getByRole('alert')).toContainText(/Duration must be greater than zero/i)
  await expect(page.locator('#activity-modal')).not.toHaveClass(/hidden/)
  await expect(page.getByRole('row').filter({ hasText: description })).toHaveCount(0)
})

test('la actividad sigue en el historial al recargar', async ({ page }) => {
  page.on('dialog', (dialog) => dialog.accept())
  await registerFreshUser(page)

  const description = `e2e reload ${Date.now()}`
  await openManualLog(page)
  await fillManualActivity(page, { description, hours: '1', minutes: '0' })
  await saveManualActivity(page)

  await expect(page.getByRole('row').filter({ hasText: description })).toBeVisible()

  await page.reload()
  await expect(page.getByRole('button', { name: 'Log Out' })).toBeVisible()
  await expect(page.getByRole('row').filter({ hasText: description })).toBeVisible()

  await page.getByRole('row').filter({ hasText: description }).getByRole('button', { name: `Delete ${description}` }).click()
  await expect(page.getByRole('row').filter({ hasText: description })).toHaveCount(0)
})
