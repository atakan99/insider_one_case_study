import { setupLayouts } from 'virtual:generated-layouts'
import { describe, expect, it } from 'vitest'
import { createMemoryHistory, createRouter } from 'vue-router'
import { routes } from 'vue-router/auto-routes'

import { resolveErrorPage } from '@/pages/errors/data/error-pages'

function makeRouter() {
  return createRouter({ history: createMemoryHistory(), routes: setupLayouts(routes) })
}

describe('routes', () => {
  it('redirects / to /dashboard', async () => {
    const router = makeRouter()
    await router.push('/')
    expect(router.currentRoute.value.path).toBe('/dashboard')
  })

  it('resolves /dashboard to the dashboard page', () => {
    const resolved = makeRouter().resolve('/dashboard')
    expect(resolved.name).toBe('/dashboard')
    for (const record of resolved.matched)
      expect([undefined, 'default']).toContain(record.meta.layout)
  })

  it.each([
    ['/campaigns/new', '/campaigns/new'],
    ['/campaigns', '/campaigns/'],
    ['/campaigns/123', '/campaigns/[id]'],
  ])('resolves the campaign page %s', (path, name) => {
    const resolved = makeRouter().resolve(path)
    expect(resolved.name).toBe(name)
    for (const record of resolved.matched)
      expect([undefined, 'default']).toContain(record.meta.layout)
  })

  it('shows 404 for unknown routes', () => {
    expect(makeRouter().resolve('/nope').name).toBe('/[...path]')
  })

  it.each(['/tasks', '/auth/sign-in', '/billing', '/settings'])(
    'shows 404 for removed route %s',
    (path) => {
      expect(makeRouter().resolve(path).name).toBe('/[...path]')
    },
  )

  it('shows 404 for removed error codes', () => {
    expect(resolveErrorPage('401').code).toBe(404)
  })
})
