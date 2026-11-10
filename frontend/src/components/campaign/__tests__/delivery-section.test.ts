// @vitest-environment happy-dom

import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import type { ChannelProgress, Progress, SendJob } from '@/lib/delivery'

import { CampaignRequestError } from '@/lib/campaigns'
import { listSendJobs, retryCampaign } from '@/lib/delivery'

import DeliverySection from '../delivery-section.vue'

vi.mock('@/lib/delivery', async (orig) => {
  const real = await orig<typeof import('@/lib/delivery')>()
  return { ...real, listSendJobs: vi.fn(), retryCampaign: vi.fn() }
})
vi.mock('vue-sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))

const zero: ChannelProgress = { total: 0, queued: 0, sending: 0, retrying: 0, sent: 0, failed: 0, unknown: 0, skipped: 0, retryable_failed: 0 }
function prog(status: Progress['status'], sms: Partial<ChannelProgress> = {}, email: Partial<ChannelProgress> = {}): Progress {
  return { campaign_id: 12, status, channels: { sms: { ...zero, ...sms }, email: { ...zero, ...email } } }
}
function job(over: Partial<SendJob> = {}): SendJob {
  return { id: 1, user_id: 4012, channel: 'sms', language: 'tr', status: 'sent', attempts: 1, last_error: '', updated_at: '2026-11-10T12:00:00Z', ...over }
}

const list = vi.mocked(listSendJobs)
const retry = vi.mocked(retryCampaign)

const completed = prog('completed', { total: 100, sent: 80, failed: 12, unknown: 5, skipped: 3, retryable_failed: 12 }, { total: 50, sent: 50 })

async function make(p: Progress = completed) {
  const w = mount(DeliverySection, { props: { campaignId: 12, progress: p }, attachTo: document.body })
  await flushPromises()
  return w
}
const $ = (sel: string) => document.body.querySelector<HTMLElement>(sel)

beforeEach(() => {
  list.mockReset()
  retry.mockReset()
  list.mockResolvedValue({ send_jobs: [job()], next_after: null })
})

afterEach(() => {
  document.body.innerHTML = ''
})

describe('delivery section', () => {
  it('shows a summary for each channel', async () => {
    const w = await make()
    const sms = w.find('[data-summary="sms"]').text()
    expect(sms).toContain('80')
    expect(sms).toContain('12')
    expect(sms).toContain('100')
    expect(w.find('[data-summary="email"]').text()).toContain('50')
  })

  it('shows View progress for approved, queued and sending only', async () => {
    for (const s of ['approved', 'queued', 'sending'] as const) {
      const w = await make(prog(s, { total: 5 }))
      expect(w.find('[data-test="view-progress"]').exists()).toBe(true)
      w.unmount()
    }
    const w = await make()
    expect(w.find('[data-test="view-progress"]').exists()).toBe(false)
  })

  it('view progress emits view-progress', async () => {
    const w = await make(prog('sending', { total: 5 }))
    await w.find('[data-test="view-progress"]').trigger('click')
    expect(w.emitted('view-progress')).toHaveLength(1)
  })

  it('shows Retry only for a completed campaign with retryable failed jobs', async () => {
    expect((await make()).find('[data-test="retry"]').text()).toBe('Retry 12 failed')
    const a = await make(prog('completed', { total: 5, sent: 5 }))
    expect(a.find('[data-test="retry"]').exists()).toBe(false)
    const b = await make(prog('sending', { total: 5, failed: 2, retryable_failed: 2 }))
    expect(b.find('[data-test="retry"]').exists()).toBe(false)
  })

  it('sums retryable failed jobs over both channels', async () => {
    const w = await make(prog('completed', { total: 5, failed: 2, retryable_failed: 2 }, { total: 5, failed: 1, retryable_failed: 1 }))
    expect(w.find('[data-test="retry"]').text()).toBe('Retry 3 failed')
  })

  it('asks before it retries, then retries and emits retried', async () => {
    const next = prog('sending', { total: 100, queued: 12 })
    retry.mockResolvedValue({ retried: 12, progress: next })
    const w = await make()
    await w.find('[data-test="retry"]').trigger('click')
    expect(retry).not.toHaveBeenCalled()
    expect($('[data-test="retry-confirm"]')!.textContent).toContain('Send 12 failed messages again? Messages with an unknown result are not sent again.')
    $('[data-test="retry-back"]')!.click()
    await flushPromises()
    expect(retry).not.toHaveBeenCalled()
    await w.find('[data-test="retry"]').trigger('click')
    expect($('[data-test="retry-confirm"]')!.textContent).toContain('Send again')
    $('[data-test="retry-send"]')!.click()
    await flushPromises()
    expect(retry).toHaveBeenCalledWith(12)
    expect(w.emitted('retried')![0]).toEqual([next])
  })

  it('shows the API message when the retry fails', async () => {
    retry.mockImplementation(() => Promise.reject(new CampaignRequestError('nothing_to_retry', 'Nothing to retry.', 409, [])))
    const w = await make()
    await w.find('[data-test="retry"]').trigger('click')
    $('[data-test="retry-send"]')!.click()
    await flushPromises()
    expect(w.emitted('retried')).toBeUndefined()
    expect(w.find('[data-test="retry-error"]').text()).toBe('Nothing to retry.')
  })

  it('lists the jobs as User #id and hints at unknown', async () => {
    list.mockResolvedValue({ send_jobs: [job(), job({ id: 2, user_id: 4031, channel: 'email', status: 'unknown', last_error: 'timeout' })], next_after: null })
    const w = await make()
    expect(list).toHaveBeenCalledWith(12, expect.objectContaining({}))
    const rows = w.findAll('[data-test="job"]')
    expect(rows).toHaveLength(2)
    expect(rows[0]!.text()).toContain('User #4012')
    expect(rows[1]!.text()).toContain('may be delivered')
    expect(rows[0]!.text()).not.toContain('may be delivered')
    expect(w.text()).not.toContain('@')
    expect(w.text()).not.toContain('+90')
  })

  it('shows the counts on the chips and filters by status', async () => {
    const w = await make()
    const chip = (k: string) => w.find(`[data-chip="${k}"]`)
    expect(chip('failed').text()).toContain('12')
    expect(chip('sent').text()).toContain('130')
    expect(chip('unknown').text()).toContain('5')
    expect(chip('skipped').text()).toContain('3')
    list.mockClear()
    await chip('failed').trigger('click')
    await flushPromises()
    expect(list).toHaveBeenCalledWith(12, expect.objectContaining({ status: 'failed' }))
    expect(chip('failed').attributes('aria-pressed')).toBe('true')
    await chip('all').trigger('click')
    await flushPromises()
    expect(list.mock.calls.at(-1)![1].status).toBeUndefined()
  })

  it('filters by channel', async () => {
    const w = await make()
    list.mockClear()
    await w.find('[data-channel-filter="email"]').trigger('click')
    await flushPromises()
    expect(list).toHaveBeenCalledWith(12, expect.objectContaining({ channel: 'email' }))
  })

  it('load more passes after = next_after and appends', async () => {
    list.mockResolvedValueOnce({ send_jobs: [job({ id: 1 })], next_after: 77 })
    const w = await make()
    list.mockResolvedValueOnce({ send_jobs: [job({ id: 2, user_id: 5000 })], next_after: null })
    await w.find('[data-test="load-more"]').trigger('click')
    await flushPromises()
    expect(list).toHaveBeenLastCalledWith(12, expect.objectContaining({ after: 77 }))
    expect(w.findAll('[data-test="job"]')).toHaveLength(2)
    expect(w.find('[data-test="load-more"]').exists()).toBe(false)
  })

  it('reloads the list when the counts change', async () => {
    const w = await make(prog('sending', { total: 100, sent: 10 }))
    list.mockClear()
    await w.setProps({ progress: prog('sending', { total: 100, sent: 20 }) })
    await flushPromises()
    expect(list).toHaveBeenCalledTimes(1)
  })

  it('shows an empty note and a list error', async () => {
    list.mockResolvedValue({ send_jobs: [], next_after: null })
    const w = await make()
    expect(w.find('[data-test="empty"]').exists()).toBe(true)
    list.mockImplementation(() => Promise.reject(new Error('x')))
    await w.find('[data-chip="failed"]').trigger('click')
    await flushPromises()
    expect(w.find('[data-test="list-error"]').text()).toBe('Something went wrong. Try again.')
  })
})
