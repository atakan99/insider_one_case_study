// @vitest-environment happy-dom

import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { Campaign } from '@/lib/campaigns'

import { CampaignRequestError, getCampaign } from '@/lib/campaigns'

import CampaignPage from '../[id].vue'

vi.mock('vue-router', () => ({ useRoute: () => ({ params: { id: '12' } }), useRouter: () => ({ push: vi.fn() }) }))
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
    expect(w.text()).toContain('Sending starts in a later step.')
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
    expect(get).toHaveBeenCalledTimes(2)
    expect(w.text()).toContain('Approved')
  })

  it('shows not found', async () => {
    get.mockImplementation(() => Promise.reject(new CampaignRequestError('not_found', 'No campaign has this id.', 404, [])))
    const w = make()
    await flushPromises()
    expect(w.text()).toContain('Campaign not found')
    expect(w.findComponent(RouterLinkStub).props('to')).toBe('/campaigns')
  })
})
