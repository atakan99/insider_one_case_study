import { useQueryClient } from '@tanstack/vue-query'
import { afterAll, afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createApp } from 'vue'

import { apiFetch } from '@/lib/api-client'
import { setupTanstackVueQuery } from '@/plugins/tanstack-vue-query/setup'

const { fetchMock, notify } = vi.hoisted(() => {
  const fetchMock = vi.fn()
  vi.stubGlobal('fetch', fetchMock)
  return { fetchMock, notify: vi.fn() }
})

vi.mock('vue-sonner', () => ({ toast: { error: notify } }))
vi.mock('@/constants/app-config', () => ({ API_BASE_URL: 'https://example.test/api', API_TIMEOUT: 5000 }))

const app = createApp({})
setupTanstackVueQuery(app)
const queryClient = app.runWithContext(() => useQueryClient())

function respond(status: number) {
  return new Response(JSON.stringify({ message: 'Service unavailable' }), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

function fetchQuery() {
  return queryClient.fetchQuery({
    queryKey: ['campaigns'],
    queryFn: () => apiFetch('/campaigns'),
    retry: 1,
    retryDelay: 0,
  })
}

describe('aPI error handling', () => {
  beforeEach(() => {
    vi.resetAllMocks()
  })

  afterEach(() => {
    queryClient.clear()
  })

  afterAll(() => {
    vi.unstubAllGlobals()
  })

  it.each([
    [503, 200],
    [503, 503, 200],
  ])('does not notify when retries recover (%j)', async (...statuses) => {
    fetchMock.mockImplementation(() => respond(statuses.shift()!))

    await expect(fetchQuery()).resolves.toBeDefined()
    expect(notify).not.toHaveBeenCalled()
  })

  it('notifies once after both fetch and query retries are exhausted', async () => {
    fetchMock.mockImplementation(() => {
      expect(notify).not.toHaveBeenCalled()
      return respond(503)
    })

    await expect(fetchQuery()).rejects.toMatchObject({ status: 503 })
    expect(fetchMock).toHaveBeenCalledTimes(4)
    expect(notify).toHaveBeenCalledExactlyOnceWith('Service unavailable')
  })

  it('notifies once after mutation retries are exhausted', async () => {
    fetchMock.mockImplementation(() => respond(503))
    const mutation = queryClient.getMutationCache().build(queryClient, {
      mutationFn: () => apiFetch('/campaigns', { method: 'post' }),
      retry: 1,
      retryDelay: 0,
    })

    await expect(mutation.execute(undefined)).rejects.toMatchObject({ status: 503 })
    expect(fetchMock).toHaveBeenCalledTimes(2)
    expect(notify).toHaveBeenCalledExactlyOnceWith('Service unavailable')
  })

  it('does not notify for a silent mutation error', async () => {
    fetchMock.mockImplementation(() => respond(503))
    const mutation = queryClient.getMutationCache().build(queryClient, {
      mutationFn: () => apiFetch('/campaigns', { method: 'post' }),
      meta: { silent: true },
    })

    await expect(mutation.execute(undefined)).rejects.toMatchObject({ status: 503 })
    expect(notify).not.toHaveBeenCalled()
  })

  it('notifies once for a 401 response', async () => {
    fetchMock.mockImplementation(() => respond(401))

    await expect(fetchQuery()).rejects.toMatchObject({ status: 401 })
    expect(notify).toHaveBeenCalledExactlyOnceWith('Service unavailable')
  })
})
