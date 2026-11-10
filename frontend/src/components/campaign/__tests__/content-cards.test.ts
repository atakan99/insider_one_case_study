// @vitest-environment happy-dom

import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { ContentEvent } from '@/lib/content'

import { ContentRequestError, generateContent } from '@/lib/content'

import ContentCards from '../content-cards.vue'

const push = vi.fn()
vi.mock('vue-router', () => ({ useRouter: () => ({ push }) }))
vi.mock('../review-dialog.vue', async () => {
  const { defineComponent, h } = await import('vue')
  return {
    default: defineComponent({
      props: { campaignId: Number, open: Boolean },
      emits: ['approved', 'update:open'],
      setup(props, { emit }) {
        return () => props.open
          ? h('div', { 'data-test': 'review-stub', 'data-id': props.campaignId }, [
              h('button', { 'data-test': 'stub-approve', 'onClick': () => emit('approved', props.campaignId) }, 'approve'),
            ])
          : null
      },
    }),
  }
})

vi.mock('@/lib/content', async () => {
  class ContentRequestError extends Error {
    code: string
    constructor(code: string, message: string) {
      super(message)
      this.code = code
    }
  }
  return { generateContent: vi.fn(), ContentRequestError }
})

type OnEvent = (e: ContentEvent) => void
const gen = vi.mocked(generateContent)
const content = { sms: 'Hi Ayse', email_subject: 'Sale', email_body: 'Body' }
const limits = { first_name_runes: 5, sms_runes: 121, email_subject_runes: 44, email_body_runes: 10 }
const audience: ContentEvent = {
  event: 'audience',
  data: {
    description: 'd',
    languages: [
      { code: 'tr', email: 100, sms: 50, reachable: 110, matched: 120, no_consent: 10 },
      { code: 'en', email: 30, sms: 20, reachable: 40, matched: 45, no_consent: 5 },
    ],
  },
}

function make(runId = 1) {
  return mount(ContentCards, { props: { brief: 'my brief', rule: { a: 1 }, runId } })
}

// Starts a request that stays open and returns its onEvent.
function open(): { get: () => OnEvent, finish: () => void } {
  let on: OnEvent = () => {}
  let finish = () => {}
  gen.mockImplementationOnce((_b, _r, onEvent) => {
    on = onEvent
    return new Promise<void>((res) => {
      finish = res
    })
  })
  return { get: () => on, finish: () => finish() }
}

describe('content cards', () => {
  beforeEach(() => gen.mockReset())

  it('starts on the run id and shows progress', async () => {
    open()
    const w = make()
    await flushPromises()
    expect(gen).toHaveBeenCalledTimes(1)
    expect(gen.mock.calls[0]![0]).toBe('my brief')
    expect(gen.mock.calls[0]![1]).toEqual({ a: 1 })
    expect(w.text()).toContain('Generating content…')
  })

  it('does not start without a run', async () => {
    make(0)
    await flushPromises()
    expect(gen).not.toHaveBeenCalled()
  })

  it('streams drafts and shows final with counters', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    await flushPromises()
    expect(w.text()).toContain('Turkish')
    expect(w.text()).toContain('English')
    expect(w.text()).toContain('100 email, 50 SMS')
    r.get()({ event: 'draft', data: { language: 'tr', content } })
    await flushPromises()
    const draft = w.find('[data-language="tr"] [data-field="sms"]')
    expect(draft.text()).toBe('Hi Ayse')
    expect(draft.element.parentElement!.parentElement!.className).toContain('italic')
    expect(w.text()).toContain('Draft')
    r.get()({ event: 'final', data: { language: 'tr', content, limits } })
    await flushPromises()
    expect(w.text()).toContain('121 / 160')
    expect(w.text()).toContain('44 / 80')
    expect(w.text()).not.toContain('Draft')
    expect(w.find('[data-language="tr"] [data-field="sms"]').element.parentElement!.parentElement!.className).not.toContain('italic')
  })

  it('shows retry and per-language error', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    r.get()({ event: 'retry', data: { language: 'tr' } })
    r.get()({ event: 'final', data: { language: 'tr', content, limits } })
    r.get()({ event: 'retry', data: { language: 'en' } })
    await flushPromises()
    expect(w.text()).toContain('Fixing content…')
    r.get()({ event: 'error', data: { language: 'en', code: 'llm_error', message: 'English failed.' } })
    await flushPromises()
    expect(w.text()).toContain('English failed.')
    expect(w.text()).not.toContain('Fixing content…')
    expect(w.find('[data-language="tr"]').text()).toContain('121 / 160')
  })

  it('clears the draft on retry and shows the next draft', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    r.get()({ event: 'draft', data: { language: 'tr', content } })
    await flushPromises()
    expect(w.find('[data-language="tr"] [data-field="sms"]').text()).toBe('Hi Ayse')
    r.get()({ event: 'retry', data: { language: 'tr' } })
    await flushPromises()
    expect(w.find('[data-language="tr"] [data-field="sms"]').exists()).toBe(false)
    expect(w.find('[data-language="tr"]').text()).not.toContain('Hi Ayse')
    expect(w.find('[data-language="tr"]').text()).toContain('Fixing content…')
    r.get()({ event: 'draft', data: { language: 'tr', content: { ...content, sms: 'Fixed SMS' } } })
    await flushPromises()
    expect(w.find('[data-language="tr"] [data-field="sms"]').text()).toBe('Fixed SMS')
  })

  it('regenerates with the same brief and rule, off while running', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    await flushPromises()
    expect(w.find('button').text()).toBe('Regenerate content')
    expect(w.find('button').attributes('disabled')).toBeDefined()
    r.get()({ event: 'done', data: {} })
    r.finish()
    await flushPromises()
    expect(w.find('button').attributes('disabled')).toBeUndefined()
    open()
    await w.find('button').trigger('click')
    await flushPromises()
    expect(gen).toHaveBeenCalledTimes(2)
    expect(gen.mock.calls[1]![0]).toBe('my brief')
    expect(gen.mock.calls[1]![1]).toEqual({ a: 1 })
    expect(w.findAll('[data-language]').length).toBe(0)
  })

  it('ignores events from an aborted stream', async () => {
    const r1 = open()
    const w = make(1)
    await flushPromises()
    const old = r1.get()
    const r2 = open()
    await w.setProps({ runId: 2 })
    await flushPromises()
    expect((gen.mock.calls[0]![3] as AbortSignal).aborted).toBe(true)
    r2.get()(audience)
    old({ event: 'final', data: { language: 'tr', content, limits } })
    old({ event: 'error', data: { language: 'en', code: 'llm_error', message: 'stale' } })
    await flushPromises()
    expect(w.text()).not.toContain('121 / 160')
    expect(w.text()).not.toContain('stale')
    expect(w.findAll('[data-language]').length).toBe(2)
  })

  it('marks open cards on early close', async () => {
    gen.mockReset()
    gen.mockImplementationOnce(async (_b, _r, onEvent) => {
      onEvent(audience)
      onEvent({ event: 'final', data: { language: 'tr', content, limits } })
      throw new ContentRequestError('stream_closed', 'The connection closed early. Try again.')
    })
    const w = make()
    await flushPromises()
    expect(w.find('[data-language="tr"]').text()).not.toContain('closed early')
    expect(w.find('[data-language="en"]').text()).toContain('The connection closed early. Try again.')
  })

  it('shows text, not HTML', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    r.get()({ event: 'draft', data: { language: 'tr', content: { ...content, sms: '<b>x</b>' } } })
    await flushPromises()
    expect(w.text()).toContain('<b>x</b>')
    expect(w.find('b').exists()).toBe(false)
  })

  it('shows an error before the stream', async () => {
    gen.mockRejectedValueOnce(new ContentRequestError('empty_audience', 'No customers match.'))
    const w = make()
    await flushPromises()
    expect(w.text()).toContain('No customers match.')
    expect(w.find('button').attributes('disabled')).toBeUndefined()
  })

  it('emits running', async () => {
    const r = open()
    const w = make()
    await flushPromises()
    r.finish()
    await flushPromises()
    expect(w.emitted('update:running')!.map(a => a[0])).toEqual([true, false])
  })

  it('review button needs a campaign event', async () => {
    const r = open()
    const w = make()
    await flushPromises()
    const btn = () => w.find('[data-test="review-button"]')
    expect(btn().exists()).toBe(false)
    r.get()(audience)
    r.get()({ event: 'final', data: { language: 'tr', content, limits } })
    r.get()({ event: 'campaign', data: { id: 7, status: 'draft' } })
    await flushPromises()
    expect(btn().attributes('disabled')).toBeDefined()
    r.get()({ event: 'done', data: {} })
    r.finish()
    await flushPromises()
    expect(btn().attributes('disabled')).toBeUndefined()
    await btn().trigger('click')
    expect(w.find('[data-test="review-stub"]').attributes('data-id')).toBe('7')
    await w.find('[data-test="stub-approve"]').trigger('click')
    expect(push).toHaveBeenCalledWith('/campaigns/7')
  })

  it('no review button without a campaign event', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    r.get()({ event: 'error', data: { language: 'tr', code: 'llm_error', message: 'x' } })
    r.get()({ event: 'done', data: {} })
    r.finish()
    await flushPromises()
    expect(w.find('[data-test="review-button"]').exists()).toBe(false)
  })

  it('regenerate sends the campaign id and a new run does not', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    r.get()({ event: 'campaign', data: { id: 7, status: 'draft' } })
    r.get()({ event: 'done', data: {} })
    r.finish()
    await flushPromises()
    expect(gen.mock.calls[0]![4]).toBeUndefined()
    open()
    await w.find('button').trigger('click')
    await flushPromises()
    expect(gen.mock.calls[1]![4]).toBe(7)
    open()
    await w.setProps({ runId: 2 })
    await flushPromises()
    expect(gen.mock.calls[2]![4]).toBeUndefined()
  })

  it('shows an error with no language in the content area', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    r.get()({ event: 'final', data: { language: 'tr', content, limits } })
    r.get()({ event: 'error', data: { code: 'draft_not_saved', message: 'The draft could not be saved. Try again.' } })
    await flushPromises()
    expect(w.find('[role="alert"], .alert, [data-slot="alert"]').text()).toContain('The draft could not be saved.')
    expect(w.find('[data-language="tr"]').exists()).toBe(true)
  })

  it('hides the review button during a regenerate and after a failed save', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    r.get()({ event: 'campaign', data: { id: 7, status: 'draft' } })
    r.get()({ event: 'done', data: {} })
    r.finish()
    await flushPromises()
    expect(w.find('[data-test="review-button"]').exists()).toBe(true)
    const r2 = open()
    await w.find('button').trigger('click')
    await flushPromises()
    expect(w.find('[data-test="review-button"]').exists()).toBe(false)
    r2.get()(audience)
    r2.get()({ event: 'error', data: { code: 'campaign_conflict', message: 'This draft cannot be updated.' } })
    r2.get()({ event: 'done', data: {} })
    r2.finish()
    await flushPromises()
    expect(w.find('[data-test="review-button"]').exists()).toBe(false)
    // After a conflict, the next regenerate starts a new draft.
    open()
    await w.find('button').trigger('click')
    await flushPromises()
    expect(gen.mock.calls[2]![4]).toBeUndefined()
  })

  it('a regenerate after a failed run still replaces the draft', async () => {
    const r = open()
    const w = make()
    r.get()(audience)
    r.get()({ event: 'campaign', data: { id: 7, status: 'draft' } })
    r.get()({ event: 'done', data: {} })
    r.finish()
    await flushPromises()
    const r2 = open()
    await w.find('button').trigger('click')
    r2.get()(audience)
    r2.get()({ event: 'error', data: { language: 'tr', code: 'llm_error', message: 'x' } })
    r2.get()({ event: 'done', data: {} })
    r2.finish()
    await flushPromises()
    open()
    await w.find('button').trigger('click')
    await flushPromises()
    expect(gen.mock.calls[2]![4]).toBe(7)
  })
})
