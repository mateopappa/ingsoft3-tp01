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
  await expect(page.getByLabel('Description')).toBeVisible()
}

test('crear una actividad la muestra en la lista, y borrarla la saca', async ({ page }) => {
  page.on('dialog', (dialog) => dialog.accept())
  await registerFreshUser(page)

  const description = `e2e ${Date.now()}`
  await openManualLog(page)
  await page.getByLabel('Description').fill(description)
  await page.getByLabel('Category').selectOption({ label: 'Work' })
  await page.getByLabel('Hours').fill('0')
  await page.getByLabel('Minutes').fill('30')
  await page.getByRole('button', { name: 'Save Activity' }).click()

  await expect(page.getByText(description)).toBeVisible()

  await page.getByRole('button', { name: `Delete ${description}` }).click()
  await expect(page.getByText(description)).toHaveCount(0)
})

test('duración cero muestra el error y no crea nada', async ({ page }) => {
  await registerFreshUser(page)

  const description = `e2e bad ${Date.now()}`
  await openManualLog(page)
  await page.getByLabel('Description').fill(description)
  await page.getByLabel('Category').selectOption({ label: 'Work' })
  await page.getByLabel('Hours').fill('0')
  await page.getByLabel('Minutes').fill('0')
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
  await page.getByLabel('Description').fill(description)
  await page.getByLabel('Category').selectOption({ label: 'Work' })
  await page.getByLabel('Hours').fill('1')
  await page.getByLabel('Minutes').fill('0')
  await page.getByRole('button', { name: 'Save Activity' }).click()
  await expect(page.getByText(description)).toBeVisible()

  await page.reload()
  await expect(page.getByRole('button', { name: 'Log Out' })).toBeVisible()
  await expect(page.getByText(description)).toBeVisible()

  await page.getByRole('button', { name: `Delete ${description}` }).click()
  await expect(page.getByText(description)).toHaveCount(0)
})
