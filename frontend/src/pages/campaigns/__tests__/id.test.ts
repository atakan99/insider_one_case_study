// @vitest-environment happy-dom

import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { Campaign } from '@/lib/campaigns'

import { CampaignRequestError, getCampaign } from '@/lib/campaigns'
import { getProgress } from '@/lib/delivery'

import CampaignPage from '../[id].vue'

const push = vi.fn()
const replace = vi.fn()
const route = vi.hoisted(() => ({ params: { id: '12' } as Record<string, string>, query: {} as Record<string, string> }))
vi.mock('vue-router', () => ({ useRoute: () => route, useRouter: () => ({ push, replace }) }))
vi.mock('@/lib/delivery', async (orig) => {
  const real = await orig<typeof import('@/lib/delivery')>()
  return { ...real, getProgress: vi.fn() }
})
vi.mock('@/components/campaign/delivery-section.vue', async () => {
  const { defineComponent, h } = await import('vue')
  return {
    default: defineComponent({
      props: { campaignId: Number, progress: Object },
      emits: ['view-progress', 'retried'],
      setup(props, { emit }) {
        return () => h('div', { 'data-test': 'delivery-stub', 'data-status': (props.progress as { status: string }).status }, [
          h('button', { 'data-test': 'stub-view', 'onClick': () => emit('view-progress') }, 'v'),
        ])
      },
    }),
  }
})
vi.mock('@/components/campaign/sending-dialog.vue', async () => {
  const { defineComponent, h } = await import('vue')
  return {
    default: defineComponent({
      props: { campaignId: Number, open: Boolean },
      emits: ['update:open', 'open-list'],
      setup(props, { emit }) {
        return () => props.open
          ? h('div', { 'data-test': 'sending-stub' }, [h('button', { 'data-test': 'stub-open-list', 'onClick': () => emit('open-list') }, 'l')])
          : null
      },
    }),
  }
})
vi.mock('@/lib/campaigns', async (orig) => {
  const real = await orig<typeof import('@/lib/campaigns')>()
  return { ...real, getCampaign: vi.fn() }
})
vi.mock('@/components/campaign/review-dialog.vue', async () => {
  const { defineComponent, h } = await import('vue')
  return {
    default: defineComponent({
      props: { campaignId: Number, open: Boolean },
      emits: ['approved', 'update:open'],
      setup(props, { emit }) {
        return () => props.open
          ? h('div', { 'data-test': 'review-stub' }, [h('button', { 'data-test': 'stub-approve', 'onClick': () => emit('approved', 12) }, 'a')])
          : null
      },
    }),
  }
})

const get = vi.mocked(getCampaign)
const progress = vi.mocked(getProgress)
const zero = { total: 0, queued: 0, sending: 0, retrying: 0, sent: 0, failed: 0, unknown: 0, skipped: 0, retryable_failed: 0 }
const prog = (status: string) => ({ campaign_id: 12, status, channels: { sms: zero, email: zero } }) as never

function campaign(over: Partial<Campaign> = {}): Campaign {
  return {
    id: 12,
    status: 'approved',
    brief: 'b',
    description: 'Istanbul users',
    skipped_languages: [],
    created_at: '2026-11-10T12:00:00Z',
    updated_at: '2026-11-10T12:00:00Z',
    approved_at: '2026-11-10T12:00:00Z',
    contents: [{ language: 'tr', sms: 'Merhaba {{first_name}}', email_subject: 'İndirim', email_body: '<b>Gövde</b>', sms_count: 980, email_count: 1212, first_name_runes: 14 }],
    ...over,
  }
}

function make() {
  return mount(CampaignPage, { global: { stubs: { RouterLink: RouterLinkStub } } })
}

describe('campaign page', () => {
  beforeEach(() => {
    get.mockReset()
    progress.mockReset()
    progress.mockResolvedValue(prog('approved'))
    push.mockReset()
    replace.mockReset()
    route.query = {}
  })

  it('shows an approved campaign read-only', async () => {
    get.mockResolvedValue(campaign())
    const w = make()
    await flushPromises()
    expect(get).toHaveBeenCalledWith(12)
    expect(w.text()).toContain('Campaign #12')
    expect(w.text()).toContain('Approved')
    expect(w.text()).toContain('Istanbul users')
    const card = w.find('[data-language="tr"]')
    expect(card.text()).toContain('Merhaba {{first_name}}')
    expect(card.text()).toContain('980 SMS · 1,212 email')
    expect(card.text()).toContain('<b>Gövde</b>')
    expect(w.find('textarea').exists()).toBe(false)
    expect(w.text()).not.toContain('Sending starts in a later step.')
    expect(progress).toHaveBeenCalledWith(12)
    expect(w.find('[data-test="delivery-stub"]').attributes('data-status')).toBe('approved')
    expect(w.find('[data-test="review-button"]').exists()).toBe(false)
  })

  it('a draft shows Review & approve and reloads after approval', async () => {
    get.mockResolvedValue(campaign({ status: 'draft', approved_at: null }))
    const w = make()
    await flushPromises()
    expect(w.text()).toContain('Draft')
    await w.find('[data-test="review-button"]').trigger('click')
    get.mockResolvedValue(campaign())
    await w.find('[data-test="stub-approve"]').trigger('click')
    await flushPromises()
    expect(push).toHaveBeenCalledWith('/campaigns/12?sending=1')
  })

  it('a draft shows no delivery section and loads no progress', async () => {
    get.mockResolvedValue(campaign({ status: 'draft', approved_at: null }))
    const w = make()
    await flushPromises()
    expect(w.find('[data-test="delivery-stub"]').exists()).toBe(false)
    expect(progress).not.toHaveBeenCalled()
  })

  it('sending=1 opens the sending window and removes the query', async () => {
    route.query = { sending: '1' }
    get.mockResolvedValue(campaign())
    const w = make()
    await flushPromises()
    expect(w.find('[data-test="sending-stub"]').exists()).toBe(true)
    expect(replace).toHaveBeenCalledWith({ query: {} })
  })

  it('without sending=1 the window is closed until View progress', async () => {
    get.mockResolvedValue(campaign({ status: 'sending' }))
    const w = make()
    await flushPromises()
    expect(w.find('[data-test="sending-stub"]').exists()).toBe(false)
    await w.find('[data-test="stub-view"]').trigger('click')
    expect(w.find('[data-test="sending-stub"]').exists()).toBe(true)
  })

  it('open delivery list scrolls to the delivery section', async () => {
    const scroll = vi.fn()
    Element.prototype.scrollIntoView = scroll
    route.query = { sending: '1' }
    get.mockResolvedValue(campaign())
    const w = mount(CampaignPage, { global: { stubs: { RouterLink: RouterLinkStub } }, attachTo: document.body })
    await flushPromises()
    await w.find('[data-test="stub-open-list"]').trigger('click')
    expect(scroll).toHaveBeenCalled()
    w.unmount()
  })

  it('shows not found', async () => {
    get.mockImplementation(() => Promise.reject(new CampaignRequestError('not_found', 'No campaign has this id.', 404, [])))
    const w = make()
    await flushPromises()
    expect(w.text()).toContain('Campaign not found')
    expect(w.findComponent(RouterLinkStub).props('to')).toBe('/campaigns')
  })
})
