// @vitest-environment happy-dom

import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import CampaignsTable from '../campaigns-table.vue'

const stubs = { RouterLink: { props: ['to'], template: '<a :href="to"><slot /></a>' } }
const row = { id: 7, description: 'Spring sale', status: 'completed' as const, total: 120, sent: 95, failed: 3, unknown: 2, duration_ms: 3120 }
const empty = { id: 8, description: 'Nothing yet', status: 'queued' as const, total: 0, sent: 0, failed: 0, unknown: 0, duration_ms: null }

describe('campaigns table', () => {
  it('shows seven columns and a link to each campaign', () => {
    const w = mount(CampaignsTable, { props: { campaigns: [row] }, global: { stubs } })
    const heads = w.findAll('thead th').map(h => h.text())
    expect(heads).toEqual(['Campaign', 'Status', 'Send jobs', 'Sent %', 'Failed %', 'Unknown %', 'Duration'])
    expect(w.find('a[href="/campaigns/7"]').exists()).toBe(true)
    const text = w.find('tbody').text()
    expect(text).toContain('Spring sale')
    expect(text).toContain('#7')
    expect(text).toContain('120')
    expect(text).toContain('79.2%')
    expect(text).toContain('2.5%')
    expect(text).toContain('1.7%')
    expect(text).toContain('3.1 s')
  })

  it('shows 0 and dashes for a campaign without send jobs', () => {
    const w = mount(CampaignsTable, { props: { campaigns: [empty] }, global: { stubs } })
    const cells = w.findAll('tbody td').map(c => c.text())
    expect(cells.slice(2)).toEqual(['0', '–', '–', '–', '–'])
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
