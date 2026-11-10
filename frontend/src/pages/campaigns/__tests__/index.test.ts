// @vitest-environment happy-dom

import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { listCampaigns } from '@/lib/campaigns'

import CampaignsPage from '../index.vue'

const push = vi.fn()
vi.mock('vue-router', () => ({ useRouter: () => ({ push }) }))
vi.mock('@/lib/campaigns', async (orig) => {
  const real = await orig<typeof import('@/lib/campaigns')>()
  return { ...real, listCampaigns: vi.fn() }
})
vi.mock('@/components/campaign/review-dialog.vue', async () => {
  const { defineComponent, h } = await import('vue')
  return {
    default: defineComponent({
      props: { campaignId: Number, open: Boolean },
      emits: ['approved', 'update:open'],
      setup(props, { emit }) {
        return () => props.open
          ? h('div', { 'data-test': 'review-stub', 'data-id': props.campaignId }, [
              h('button', { 'data-test': 'stub-approve', 'onClick': () => emit('approved', props.campaignId) }, 'approve'),
              h('button', { 'data-test': 'stub-close', 'onClick': () => emit('update:open', false) }, 'close'),
            ])
          : null
      },
    }),
  }
})

const list = vi.mocked(listCampaigns)

function make() {
  return mount(CampaignsPage, { global: { stubs: { RouterLink: RouterLinkStub } } })
}

describe('campaigns page', () => {
  beforeEach(() => {
    list.mockReset()
    push.mockReset()
  })

  it('lists campaigns with status badges', async () => {
    list.mockResolvedValue([
      { id: 2, status: 'draft', description: 'Istanbul users', languages: ['en', 'tr'], created_at: '2026-11-10T12:00:00Z', approved_at: null },
      { id: 1, status: 'approved', description: 'Ankara users', languages: ['tr'], created_at: '2026-11-09T12:00:00Z', approved_at: '2026-11-09T13:00:00Z' },
    ])
    const w = make()
    await flushPromises()
    const rows = w.findAll('[data-test="campaign-row"]')
    expect(rows).toHaveLength(2)
    expect(rows[0]!.text()).toContain('#2')
    expect(rows[0]!.text()).toContain('Istanbul users')
    expect(rows[0]!.text()).toContain('English, Turkish')
    expect(rows[0]!.text()).toContain('Draft')
    expect(rows[1]!.text()).toContain('Approved')
  })

  it('a draft opens the review window, an approved campaign opens its page', async () => {
    list.mockResolvedValue([
      { id: 2, status: 'draft', description: 'd', languages: ['tr'], created_at: '2026-11-10T12:00:00Z', approved_at: null },
      { id: 1, status: 'approved', description: 'a', languages: ['tr'], created_at: '2026-11-09T12:00:00Z', approved_at: '2026-11-09T13:00:00Z' },
    ])
    const w = make()
    await flushPromises()
    await w.findAll('[data-test="campaign-row"]')[0]!.trigger('click')
    expect(w.find('[data-test="review-stub"]').attributes('data-id')).toBe('2')
    await w.findAll('[data-test="campaign-row"]')[1]!.trigger('click')
    expect(push).toHaveBeenCalledWith('/campaigns/1')
  })

  it('reloads after approval and goes to the campaign', async () => {
    list.mockResolvedValue([{ id: 2, status: 'draft', description: 'd', languages: ['tr'], created_at: '2026-11-10T12:00:00Z', approved_at: null }])
    const w = make()
    await flushPromises()
    await w.find('[data-test="campaign-row"]').trigger('click')
    await w.find('[data-test="stub-approve"]').trigger('click')
    await flushPromises()
    expect(push).toHaveBeenCalledWith('/campaigns/2')
  })

  it('reloads after the window closes', async () => {
    list.mockResolvedValue([{ id: 2, status: 'draft', description: 'd', languages: ['tr'], created_at: '2026-11-10T12:00:00Z', approved_at: null }])
    const w = make()
    await flushPromises()
    await w.find('[data-test="campaign-row"]').trigger('click')
    await w.find('[data-test="stub-close"]').trigger('click')
    await flushPromises()
    expect(list).toHaveBeenCalledTimes(2)
  })

  it('shows the empty state', async () => {
    list.mockResolvedValue([])
    const w = make()
    await flushPromises()
    expect(w.text()).toContain('No campaigns yet')
    expect(w.findComponent(RouterLinkStub).props('to')).toBe('/campaigns/new')
  })

  it('shows a load error', async () => {
    list.mockImplementation(() => Promise.reject(new Error('x')))
    const w = make()
    await flushPromises()
    expect(w.text()).toContain('Something went wrong. Try again.')
  })
})
