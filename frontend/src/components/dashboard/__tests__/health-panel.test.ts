// @vitest-environment happy-dom

import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import HealthPanel from '../health-panel.vue'

const health = {
  services: { db: 'ok', redis: 'down', worker: 'ok', provider: 'ok' } as const,
  streams: [
    { channel: 'sms' as const, length: 12, pending: 3, in_progress: 2 },
    { channel: 'email' as const, length: null, pending: null, in_progress: 0 },
  ],
}

describe('health panel', () => {
  it('shows four service badges', () => {
    const w = mount(HealthPanel, { props: { health } })
    expect(w.findAll('[data-service]')).toHaveLength(4)
    expect(w.find('[data-service="redis"]').attributes('data-state')).toBe('down')
    expect(w.find('[data-service="db"]').attributes('data-state')).toBe('ok')
  })

  it('shows two stream rows and a dash for null', () => {
    const w = mount(HealthPanel, { props: { health } })
    expect(w.findAll('[data-stream]')).toHaveLength(2)
    expect(w.find('[data-stream="sms"]').text()).toContain('12')
    expect(w.find('[data-stream="email"]').text()).toContain('–')
  })
})
