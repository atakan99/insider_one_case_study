// @vitest-environment happy-dom

import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'

import type { ChannelProgress, Progress } from '@/lib/delivery'

import SendingDialog from '../sending-dialog.vue'

const zero: ChannelProgress = { total: 0, queued: 0, sending: 0, retrying: 0, sent: 0, failed: 0, unknown: 0, skipped: 0, retryable_failed: 0 }
function prog(status: Progress['status'], sms: Partial<ChannelProgress> = {}, email: Partial<ChannelProgress> = {}): Progress {
  return { campaign_id: 12, status, channels: { sms: { ...zero, ...sms }, email: { ...zero, ...email } } }
}

async function make(p: Progress | null = null) {
  const w = mount(SendingDialog, { props: { 'campaignId': 12, 'open': true, 'progress': p, 'onUpdate:open': (v: boolean) => w.setProps({ open: v }) }, attachTo: document.body })
  await flushPromises()
  return w
}
async function send(w: Awaited<ReturnType<typeof make>>, p: Progress) {
  await w.setProps({ progress: p })
  await flushPromises()
}

const $ = (sel: string) => document.body.querySelector<HTMLElement>(sel)
const $$ = (sel: string) => [...document.body.querySelectorAll<HTMLElement>(sel)]
const text = (sel: string) => $(sel)?.textContent?.trim() ?? ''
const states = () => Object.fromEntries($$('[data-step]').map(e => [e.dataset.step, e.dataset.state]))

afterEach(() => {
  document.body.innerHTML = ''
})

describe('sending dialog', () => {
  it('shows the title and the background text', async () => {
    await make()
    expect(document.body.textContent).toContain('Sending campaign #12')
    expect(document.body.textContent).toContain('You can close this. Sending continues in the background.')
  })

  it('shows Approved before the plan', async () => {
    const w = await make()
    await send(w, prog('approved'))
    expect(states()).toEqual({ approved: 'current', queued: 'todo', sending: 'todo', done: 'todo' })
    expect($$('[data-channel]')).toHaveLength(0)
    expect($('[data-test="done-text"]')).toBeNull()
  })

  it('marks the steps', async () => {
    const w = await make()
    await send(w, prog('queued', { total: 5, queued: 5 }))
    expect(states()).toEqual({ approved: 'done', queued: 'current', sending: 'todo', done: 'todo' })
    await send(w, prog('sending', { total: 5, sent: 2, queued: 3 }))
    expect(states()).toEqual({ approved: 'done', queued: 'done', sending: 'current', done: 'todo' })
    expect($('[data-step="approved"]')!.textContent).toContain('✓')
    await send(w, prog('completed', { total: 5, sent: 5 }))
    expect(states()).toEqual({ approved: 'done', queued: 'done', sending: 'done', done: 'done' })
  })

  it('sizes the bar segments by count over total', async () => {
    const w = await make()
    await send(w, prog('sending', { total: 100, sent: 50, retrying: 10, failed: 20, unknown: 20 }))
    const width = (kind: string) => $(`[data-channel="sms"] [data-segment="${kind}"]`)!.style.width
    expect(width('sent')).toBe('50%')
    expect(width('retrying')).toBe('10%')
    expect(width('failed')).toBe('20%')
    expect(width('unknown')).toBe('20%')
    expect(text('[data-channel="sms"]')).toContain('50 / 100')
  })

  it('hides a channel with no recipients', async () => {
    const w = await make()
    await send(w, prog('sending', { total: 1200, sent: 600 }, { total: 0 }))
    expect($$('[data-channel]').map(e => e.dataset.channel)).toEqual(['sms'])
    expect(text('[data-channel="sms"]')).toContain('600 / 1,200')
  })

  it('names skipped messages in the legend text, not in the bar', async () => {
    const w = await make()
    await send(w, prog('sending', { total: 10, sent: 5, skipped: 5 }))
    expect($$('[data-segment]').map(e => e.dataset.segment)).not.toContain('skipped')
    expect(text('[data-test="legend"]')).toContain('Skipped 5')
  })

  it('shows the done text at completed', async () => {
    const w = await make()
    await send(w, prog('completed', { total: 1500, sent: 1400, failed: 7, unknown: 3 }, { total: 800, sent: 770, failed: 5, unknown: 7 }))
    expect(text('[data-test="done-text"]')).toBe('Done. 2,170 sent, 12 failed, 10 unknown.')
  })

  it('open delivery list emits open-list and closes', async () => {
    const w = await make()
    await send(w, prog('sending', { total: 5 }))
    $('[data-test="open-list"]')!.click()
    await flushPromises()
    expect(w.emitted('openList')).toHaveLength(1)
    expect(w.emitted('update:open')!.at(-1)).toEqual([false])
  })

  it('close closes', async () => {
    const w = await make()
    expect(text('[data-test="close"]')).toBe('Close')
    $('[data-test="close"]')!.click()
    await flushPromises()
    expect(w.emitted('update:open')!.at(-1)).toEqual([false])
  })
})
