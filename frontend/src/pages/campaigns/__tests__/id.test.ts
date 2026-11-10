// @vitest-environment happy-dom

import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import type { Campaign } from '@/lib/campaigns'
import type { Progress } from '@/lib/delivery'

import { CampaignRequestError, getCampaign } from '@/lib/campaigns'
import { getProgress, watchProgress } from '@/lib/delivery'

import CampaignPage from '../[id].vue'

const push = vi.fn()
const replace = vi.fn()
const holder = vi.hoisted(() => ({ route: null as unknown as { params: Record<string, string>, query: Record<string, string> } }))
vi.mock('vue-router', async () => {
  const { reactive } = await import('vue')
  holder.route = reactive({ params: { id: '12' }, query: {} })
  return { useRoute: () => holder.route, useRouter: () => ({ push, replace }) }
})
const route = holder.route
vi.mock('@/lib/delivery', async (orig) => {
  const real = await orig<typeof import('@/lib/delivery')>()
  return { ...real, getProgress: vi.fn(), watchProgress: vi.fn() }
})
vi.mock('@/components/campaign/delivery-section.vue', async () => {
  const { defineComponent, h } = await import('vue')
  return {
    default: defineComponent({
      props: { campaignId: Number, progress: Object },
      emits: ['viewProgress', 'retried', 'refresh'],
      setup(props, { emit }) {
        return () => h('div', { 'data-test': 'delivery-stub', 'data-status': (props.progress as { status: string }).status }, [
          h('button', { 'data-test': 'stub-view', 'onClick': () => emit('viewProgress') }, 'v'),
          h('button', { 'data-test': 'stub-refresh', 'onClick': () => emit('refresh') }, 'r'),
          h('button', { 'data-test': 'stub-retried', 'onClick': () => emit('retried', { campaign_id: 12, status: 'sending', channels: { sms: {}, email: {} } }) }, 't'),
          (props.progress as { status: string }).status === 'completed' ? h('span', { 'data-test': 'retry-stub' }, 'Retry') : null,
          ['approved', 'queued', 'sending'].includes((props.progress as { status: string }).status) ? h('span', { 'data-test': 'view-stub' }, 'View progress') : null,
        ])
      },
    }),
  }
})
vi.mock('@/components/campaign/sending-dialog.vue', async () => {
  const { defineComponent, h } = await import('vue')
  return {
    default: defineComponent({
      props: { campaignId: Number, open: Boolean, progress: Object },
      emits: ['update:open', 'openList'],
      setup(props, { emit }) {
        return () => props.open
          ? h('div', { 'data-test': 'sending-stub', 'data-status': (props.progress as { status: string } | null)?.status }, [h('button', { 'data-test': 'stub-open-list', 'onClick': () => emit('openList') }, 'l')])
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
const watch = vi.mocked(watchProgress)
const stopWatch = vi.fn()
let pushProgress: (p: Progress) => void
const zero = { total: 0, queued: 0, sending: 0, retrying: 0, sent: 0, failed: 0, unknown: 0, skipped: 0, retryable_failed: 0 }
const prog = (status: string): Progress => ({ campaign_id: 12, status: status as Progress['status'], channels: { sms: zero, email: zero } }) as Progress

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

const mounted: { unmount: () => void }[] = []
function make() {
  const w = mount(CampaignPage, { global: { stubs: { RouterLink: RouterLinkStub } } })
  mounted.push(w)
  return w
}

describe('campaign page', () => {
  beforeEach(() => {
    get.mockReset()
    progress.mockReset()
    watch.mockReset()
    stopWatch.mockReset()
    watch.mockImplementation((_id, on) => {
      pushProgress = on
      return stopWatch
    })
    progress.mockResolvedValue(prog('approved'))
    push.mockReset()
    replace.mockReset()
    route.query = {}
    route.params.id = '12'
  })

  afterEach(() => {
    // The route is shared and reactive, so a page left mounted would react to the next test.
    for (const w of mounted.splice(0)) w.unmount()
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
    expect(w.find('[data-test="sending-stub"]').attributes('data-status')).toBe('approved')
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

  it('follows the stream: completed shows Retry and the Completed badge', async () => {
    get.mockResolvedValue(campaign({ status: 'sending' }))
    progress.mockResolvedValue(prog('sending'))
    const w = make()
    await flushPromises()
    expect(watch).toHaveBeenCalledWith(12, expect.any(Function))
    expect(w.text()).toContain('Sending')
    expect(w.find('[data-test="view-stub"]').exists()).toBe(true)
    expect(w.find('[data-test="retry-stub"]').exists()).toBe(false)
    pushProgress({ campaign_id: 12, status: 'completed', channels: { sms: { ...zero, total: 5, failed: 2, retryable_failed: 2 }, email: zero } })
    await flushPromises()
    expect(w.find('[data-status="completed"]').text()).toBe('Completed')
    expect(w.find('[data-test="retry-stub"]').exists()).toBe(true)
    expect(w.find('[data-test="view-stub"]').exists()).toBe(false)
    expect(stopWatch).toHaveBeenCalled()
  })

  it('does not watch a completed campaign and stops on unmount', async () => {
    get.mockResolvedValue(campaign({ status: 'completed' }))
    progress.mockResolvedValue(prog('completed'))
    const w = make()
    await flushPromises()
    expect(watch).not.toHaveBeenCalled()
    get.mockResolvedValue(campaign({ status: 'sending' }))
    progress.mockResolvedValue(prog('sending'))
    const w2 = make()
    await flushPromises()
    stopWatch.mockClear()
    w2.unmount()
    expect(stopWatch).toHaveBeenCalled()
    w.unmount()
  })

  it('a retry starts the watch again', async () => {
    get.mockResolvedValue(campaign({ status: 'completed' }))
    progress.mockResolvedValue(prog('completed'))
    const w = make()
    await flushPromises()
    await w.find('[data-test="stub-retried"]').trigger('click')
    expect(watch).toHaveBeenCalledTimes(1)
    expect(w.find('[data-test="sending-stub"]').exists()).toBe(true)
  })

  it('refreshes the progress when the section asks', async () => {
    get.mockResolvedValue(campaign({ status: 'completed' }))
    progress.mockResolvedValue(prog('completed'))
    const w = make()
    await flushPromises()
    progress.mockClear()
    progress.mockResolvedValue(prog('sending'))
    await w.find('[data-test="stub-refresh"]').trigger('click')
    await flushPromises()
    expect(progress).toHaveBeenCalledTimes(1)
    expect(w.find('[data-test="view-stub"]').exists()).toBe(true)
  })

  it('a progress failure shows an inline error and keeps the page', async () => {
    get.mockResolvedValue(campaign())
    progress.mockImplementation(() => Promise.reject(new CampaignRequestError('request_failed', 'Progress is not available.', 500, [])))
    const w = make()
    await flushPromises()
    expect(w.text()).toContain('Campaign #12')
    expect(w.text()).toContain('Istanbul users')
    expect(w.find('[data-test="progress-error"]').text()).toBe('Progress is not available.')
  })

  it('drops a stale progress response after the route id changes', async () => {
    let resolveOld!: (p: Progress) => void
    get.mockImplementation(async (cid: number) => campaign({ id: cid, status: 'sending' }))
    progress.mockImplementation((cid: number) => cid === 12
      ? new Promise<Progress>((r) => {
          resolveOld = r
        })
      : Promise.resolve({ ...prog('queued'), campaign_id: 13 }))
    const w = make()
    await flushPromises()
    route.params.id = '13'
    await flushPromises()
    expect(w.find('[data-test="delivery-stub"]').attributes('data-status')).toBe('queued')
    expect(watch).toHaveBeenCalledTimes(1)
    expect(watch).toHaveBeenLastCalledWith(13, expect.any(Function))
    resolveOld({ ...prog('completed'), campaign_id: 12 })
    await flushPromises()
    expect(w.find('[data-test="delivery-stub"]').attributes('data-status')).toBe('queued')
    expect(watch).toHaveBeenCalledTimes(1)
  })

  it('drops a stale progress error after the route id changes', async () => {
    let rejectOld!: (e: unknown) => void
    get.mockImplementation(async (cid: number) => campaign({ id: cid, status: 'sending' }))
    progress.mockImplementation((cid: number) => cid === 12
      ? new Promise<Progress>((_, rej) => {
          rejectOld = rej
        })
      : Promise.resolve({ ...prog('queued'), campaign_id: 13 }))
    const w = make()
    await flushPromises()
    route.params.id = '13'
    await flushPromises()
    rejectOld(new CampaignRequestError('request_failed', 'Old failure.', 500, []))
    await flushPromises()
    expect(w.find('[data-test="progress-error"]').exists()).toBe(false)
  })
})
