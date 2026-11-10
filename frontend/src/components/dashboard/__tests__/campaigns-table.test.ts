// @vitest-environment happy-dom

import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import CampaignsTable from '../campaigns-table.vue'

const stubs = { RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' } }
const row = { id: 7, description: 'Spring sale', status: 'completed' as const, total: 100, sent: 95, failed: 3, unknown: 2, duration_ms: 3120 }

describe('campaigns table', () => {
  it('shows seven columns and a link to each campaign', () => {
    const w = mount(CampaignsTable, { props: { campaigns: [row] }, global: { stubs } })
    expect(w.findAll('thead th')).toHaveLength(7)
    expect(w.find('a[href="/campaigns/7"]').exists()).toBe(true)
    const text = w.find('tbody').text()
    expect(text).toContain('Spring sale')
    expect(text).toContain('95.0%')
    expect(text).toContain('3.1 s')
  })

  it('scrolls only inside the card', () => {
    const w = mount(CampaignsTable, { props: { campaigns: [row] }, global: { stubs } })
    expect(w.find('.overflow-x-auto').exists()).toBe(true)
  })

  it('shows an empty state with a link to create a campaign', () => {
    const w = mount(CampaignsTable, { props: { campaigns: [] }, global: { stubs } })
    expect(w.text()).toContain('No campaigns sent yet')
    expect(w.find('a[href="/campaigns/new"]').exists()).toBe(true)
  })
})
