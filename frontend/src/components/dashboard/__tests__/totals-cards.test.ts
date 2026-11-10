// @vitest-environment happy-dom

import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import TotalsCards from '../totals-cards.vue'

const t = { sent: 90, failed: 3, unknown: 2, skipped: 4, in_progress: 1 }

describe('totals cards', () => {
  it('shows six cards for each channel', () => {
    const w = mount(TotalsCards, { props: { totals: { sms: t, email: t } } })
    expect(w.findAll('[data-row="sms"] [data-card]')).toHaveLength(6)
    expect(w.findAll('[data-row="email"] [data-card]')).toHaveLength(6)
  })

  it('explains that Unknown may be delivered', () => {
    const w = mount(TotalsCards, { props: { totals: { sms: t, email: t } } })
    expect(w.find('[data-row="sms"] [data-card="unknown"]').text()).toContain('may be delivered')
    expect(w.find('[data-row="sms"] [data-card="sent"]').text()).toContain('90')
  })
})
