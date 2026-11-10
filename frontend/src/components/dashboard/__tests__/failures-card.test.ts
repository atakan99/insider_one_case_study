// @vitest-environment happy-dom

import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import FailuresCard from '../failures-card.vue'

const failures = [
  { status: 'failed' as const, reason: 'provider_error', sms: 3, email: 1 },
  { status: 'unknown' as const, reason: 'timeout', sms: 2, email: 0 },
]

describe('failures card', () => {
  it('shows one row per failure with both counts', () => {
    const w = mount(FailuresCard, { props: { failures, meanAttempts: { sms: 1.04, email: null } } })
    const rows = w.findAll('[data-test="failure-row"]')
    expect(rows).toHaveLength(2)
    const cells = rows.map(r => r.findAll('td').map(c => c.text()))
    expect(cells).toEqual([
      ['failed', 'provider_error', '3', '1'],
      ['unknown', 'timeout', '2', '0'],
    ])
  })

  it('shows mean attempts and a dash for null', () => {
    const w = mount(FailuresCard, { props: { failures, meanAttempts: { sms: 1.04, email: null } } })
    expect(w.find('[data-test="mean-sms"]').text()).toContain('1.04')
    expect(w.find('[data-test="mean-email"]').text()).toContain('–')
  })
})
