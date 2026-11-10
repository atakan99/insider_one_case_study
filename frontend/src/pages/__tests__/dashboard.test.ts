// @vitest-environment happy-dom

import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import type { DashboardHealth, DashboardStats } from '@/lib/dashboard'

import { getDashboardHealth, getDashboardStats } from '@/lib/dashboard'

import Dashboard from '../dashboard.vue'

vi.mock('@/lib/dashboard', async (orig) => {
  const real = await orig<typeof import('@/lib/dashboard')>()
  return { ...real, getDashboardStats: vi.fn(), getDashboardHealth: vi.fn() }
})

const z = { sent: 0, failed: 0, unknown: 0, skipped: 0, in_progress: 0 }
const stats: DashboardStats = {
  generated_at: '2026-11-10T12:00:00Z',
  totals: { sms: { ...z, sent: 11 }, email: z },
  throughput: { bucket_seconds: 5, start: '2026-11-10T11:50:00Z', sms: Array.from<number>({ length: 120 }).fill(0), email: Array.from<number>({ length: 120 }).fill(0) },
  failures: [],
  mean_attempts: { sms: null, email: null },
  campaigns: [],
}
const health: DashboardHealth = {
  services: { db: 'ok', redis: 'ok', worker: 'ok', provider: 'ok' },
  streams: [{ channel: 'sms', length: 0, pending: 0, in_progress: 0 }, { channel: 'email', length: 0, pending: 0, in_progress: 0 }],
}

const getStats = vi.mocked(getDashboardStats)
const getHealth = vi.mocked(getDashboardHealth)
const stubs = { RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' }, ThroughputChart: true }

function setVisibility(state: 'visible' | 'hidden') {
  Object.defineProperty(document, 'visibilityState', { configurable: true, get: () => state })
}
function make() {
  return mount(Dashboard, { global: { stubs } })
}

beforeEach(() => {
  vi.useFakeTimers()
  setVisibility('visible')
  getStats.mockReset().mockResolvedValue(stats)
  getHealth.mockReset().mockResolvedValue(health)
})
afterEach(() => {
  vi.useRealTimers()
})

describe('dashboard page', () => {
  it('polls both endpoints at once and every 5 seconds', async () => {
    const w = make()
    await flushPromises()
    expect(getStats).toHaveBeenCalledTimes(1)
    expect(getHealth).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(5000)
    expect(getStats).toHaveBeenCalledTimes(2)
    expect(getHealth).toHaveBeenCalledTimes(2)
    await vi.advanceTimersByTimeAsync(5000)
    expect(getStats).toHaveBeenCalledTimes(3)
    w.unmount()
  })

  it('makes no request while the tab is hidden, and polls at once when it shows', async () => {
    setVisibility('hidden')
    const w = make()
    await flushPromises()
    await vi.advanceTimersByTimeAsync(20000)
    expect(getStats).not.toHaveBeenCalled()
    expect(getHealth).not.toHaveBeenCalled()
    setVisibility('visible')
    document.dispatchEvent(new Event('visibilitychange'))
    await flushPromises()
    expect(getStats).toHaveBeenCalledTimes(1)
    expect(getHealth).toHaveBeenCalledTimes(1)
    w.unmount()
  })

  it('makes no request when the timer fires in a hidden tab, and resumes on show', async () => {
    const w = make()
    await flushPromises()
    expect(getStats).toHaveBeenCalledTimes(1)
    setVisibility('hidden')
    await vi.advanceTimersByTimeAsync(5000)
    await vi.advanceTimersByTimeAsync(20000)
    expect(getStats).toHaveBeenCalledTimes(1)
    expect(getHealth).toHaveBeenCalledTimes(1)
    setVisibility('visible')
    document.dispatchEvent(new Event('visibilitychange'))
    await flushPromises()
    expect(getStats).toHaveBeenCalledTimes(2)
    await vi.advanceTimersByTimeAsync(5000)
    expect(getStats).toHaveBeenCalledTimes(3)
    w.unmount()
  })

  it('keeps the last data and shows Could not refresh', async () => {
    const w = make()
    await flushPromises()
    expect(w.text()).not.toContain('Could not refresh')
    getStats.mockRejectedValue(new Error('down'))
    await vi.advanceTimersByTimeAsync(5000)
    expect(w.text()).toContain('Could not refresh. Last update')
    expect(w.find('[data-row="sms"] [data-card="sent"]').text()).toContain('11')
    w.unmount()
  })

  it('does not start a second poll while one is in flight', async () => {
    let release: (v: DashboardStats) => void = () => {}
    getStats.mockReturnValue(new Promise((r) => {
      release = r
    }))
    const w = make()
    await flushPromises()
    document.dispatchEvent(new Event('visibilitychange'))
    document.dispatchEvent(new Event('visibilitychange'))
    await vi.advanceTimersByTimeAsync(20000)
    expect(getStats).toHaveBeenCalledTimes(1)
    release(stats)
    await flushPromises()
    w.unmount()
  })

  it('stops polling after unmount, also with a request in flight', async () => {
    let release: (v: DashboardStats) => void = () => {}
    getStats.mockReturnValue(new Promise((r) => {
      release = r
    }))
    const w = make()
    await flushPromises()
    const before = w.html()
    w.unmount()
    release(stats)
    await flushPromises()
    expect(w.html()).toBe(before)
    expect(w.text()).not.toContain('11')
    await vi.advanceTimersByTimeAsync(30000)
    document.dispatchEvent(new Event('visibilitychange'))
    await flushPromises()
    expect(getStats).toHaveBeenCalledTimes(1)
    expect(getHealth).toHaveBeenCalledTimes(1)
  })
})
