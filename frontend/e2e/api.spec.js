import { expect, test } from '@playwright/test'

// La api tiene su propia dirección: NO es la del front. Sin barra al final.
const API = (process.env.API_BASE_URL || 'http://localhost:8080').replace(/\/$/, '')

async function registerUser(request) {
  const email = `api-${Date.now()}-${Math.floor(Math.random() * 1e6)}@flow.test`
  const password = 'password123'
  const r = await request.post(`${API}/api/auth/register`, {
    data: { email, password },
  })
  expect(r.status()).toBe(201)
  const body = await r.json()
  expect(body.user).toBeTruthy()
  return { email, password, user: body.user }
}

async function listCategories(request) {
  const r = await request.get(`${API}/api/categories`)
  expect(r.status()).toBe(200)
  const body = await r.json()
  return body.categories || []
}

async function listActivities(request) {
  const r = await request.get(`${API}/api/activities`)
  expect(r.status()).toBe(200)
  const body = await r.json()
  return body.activities || []
}

test('el alta guarda en la base de verdad, y el borrado la saca', async ({ request }) => {
  await registerUser(request)
  const categories = await listCategories(request)
  expect(categories.length).toBeGreaterThan(0)
  const categoryId = categories[0].id

  const description = `api ${Date.now()}`
  const alta = await request.post(`${API}/api/activities`, {
    data: {
      description,
      category_id: categoryId,
      activity_date: new Date().toISOString().slice(0, 10),
      duration_seconds: 1800,
      favorite: false,
    },
  })
  expect(alta.status()).toBe(201)
  const created = await alta.json()
  const activityId = created.activity.id

  const lista = await listActivities(request)
  expect(lista.some((a) => a.description === description)).toBe(true)

  const borrado = await request.delete(`${API}/api/activities/${activityId}`)
  expect(borrado.status()).toBe(204)

  const despues = await listActivities(request)
  expect(despues.some((a) => a.description === description)).toBe(false)
})

test('una descripción vacía la rechaza la api, y no crea nada', async ({ request }) => {
  await registerUser(request)
  const categories = await listCategories(request)
  const categoryId = categories[0].id
  const antes = await listActivities(request)

  const alta = await request.post(`${API}/api/activities`, {
    data: {
      description: '',
      category_id: categoryId,
      activity_date: new Date().toISOString().slice(0, 10),
      duration_seconds: 600,
      favorite: false,
    },
  })
  expect(alta.status()).toBe(400)

  const despues = await listActivities(request)
  expect(despues.length).toBe(antes.length)
})

test('marcar favorito persiste en la base de verdad', async ({ request }) => {
  await registerUser(request)
  const categories = await listCategories(request)
  const categoryId = categories[0].id
  const description = `fav ${Date.now()}`

  const alta = await request.post(`${API}/api/activities`, {
    data: {
      description,
      category_id: categoryId,
      activity_date: new Date().toISOString().slice(0, 10),
      duration_seconds: 900,
      favorite: false,
    },
  })
  expect(alta.status()).toBe(201)
  const activityId = (await alta.json()).activity.id

  const patch = await request.patch(`${API}/api/activities/${activityId}/favorite`, {
    data: { favorite: true },
  })
  expect(patch.status()).toBe(200)

  const lista = await listActivities(request)
  const found = lista.find((a) => a.id === activityId)
  expect(found).toBeTruthy()
  expect(found.favorite).toBe(true)

  const borrado = await request.delete(`${API}/api/activities/${activityId}`)
  expect(borrado.status()).toBe(204)
})
