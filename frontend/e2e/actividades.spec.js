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
  return { email, password }
}

async function openManualLog(page) {
  await page.getByRole('button', { name: '+ Log Activity Manually' }).click()
  // Exact label: getByLabel('Description') also matches "Timer description" (strict mode).
  await expect(page.getByLabel('Description *')).toBeVisible()
  await expect(page.getByLabel('Category *').locator('option', { hasText: 'Work' })).toHaveCount(1)
}

async function fillManualActivity(page, { description, hours, minutes }) {
  await page.getByLabel('Description *').fill(description)
  await page.getByLabel('Category *').selectOption({ label: 'Work' })
  await page.getByLabel('Hours').fill(hours)
  await page.getByLabel('Minutes').fill(minutes)
}

test('crear una actividad la muestra en la lista, y borrarla la saca', async ({ page }) => {
  page.on('dialog', (dialog) => dialog.accept())
  await registerFreshUser(page)

  const description = `e2e ${Date.now()}`
  await openManualLog(page)
  await fillManualActivity(page, { description, hours: '0', minutes: '30' })
  await page.getByRole('button', { name: 'Save Activity' }).click()

  await expect(page.getByText(description)).toBeVisible()

  await page.getByRole('button', { name: `Delete ${description}` }).click()
  await expect(page.getByText(description)).toHaveCount(0)
})

test('duración cero muestra el error y no crea nada', async ({ page }) => {
  await registerFreshUser(page)

  const description = `e2e bad ${Date.now()}`
  await openManualLog(page)
  await fillManualActivity(page, { description, hours: '0', minutes: '0' })
  await page.getByRole('button', { name: 'Save Activity' }).click()

  await expect(page.getByRole('alert')).toBeVisible()
  await expect(page.getByRole('alert')).toContainText(/Duration must be greater than zero/i)
  await expect(page.getByText(description)).toHaveCount(0)
})

test('la actividad sigue en el historial al recargar', async ({ page }) => {
  page.on('dialog', (dialog) => dialog.accept())
  await registerFreshUser(page)

  const description = `e2e reload ${Date.now()}`
  await openManualLog(page)
  await fillManualActivity(page, { description, hours: '1', minutes: '0' })
  await page.getByRole('button', { name: 'Save Activity' }).click()
  await expect(page.getByText(description)).toBeVisible()

  await page.reload()
  await expect(page.getByRole('button', { name: 'Log Out' })).toBeVisible()
  await expect(page.getByText(description)).toBeVisible()

  await page.getByRole('button', { name: `Delete ${description}` }).click()
  await expect(page.getByText(description)).toHaveCount(0)
})
