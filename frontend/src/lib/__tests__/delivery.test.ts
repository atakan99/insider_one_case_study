import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

import type { Progress } from '@/lib/delivery'

import { API_BASE_URL } from '@/constants/app-config'
import { apiFetch } from '@/lib/api-client'
import { currentStep, getProgress, listSendJobs, retryCampaign, watchProgress } from '@/lib/delivery'

vi.mock('@/lib/api-client', () => ({ apiFetch: vi.fn() }))

const zero = { total: 0, queued: 0, sending: 0, retrying: 0, sent: 0, failed: 0, unknown: 0, skipped: 0, retryable_failed: 0 }
function progress(status: Progress['status'] = 'sending', sent = 0): Progress {
  return { campaign_id: 12, status, channels: { sms: { ...zero, total: 10, sent }, email: { ...zero } } }
}

class FakeEventSource {
  static instances: FakeEventSource[] = []
  url: string
  closed = false
  listeners: Record<string, ((e: MessageEvent) => void)[]> = {}
  constructor(url: string) {
    this.url = url
    FakeEventSource.instances.push(this)
  }

  addEventListener(name: string, fn: (e: MessageEvent) => void) {
    (this.listeners[name] ??= []).push(fn)
  }

  close() {
    this.closed = true
  }

  emit(name: string, data?: unknown) {
    for (const fn of this.listeners[name] ?? [])
      fn({ data: data === undefined ? '' : JSON.stringify(data) } as MessageEvent)
  }
}

beforeEach(() => {
  FakeEventSource.instances = []
  vi.stubGlobal('EventSource', FakeEventSource)
  vi.mocked(apiFetch).mockReset()
})

afterEach(() => {
  vi.useRealTimers()
  vi.unstubAllGlobals()
})

describe('currentStep', () => {
  it('maps each status to a step', () => {
    expect(currentStep('approved')).toBe('approved')
    expect(currentStep('queued')).toBe('queued')
    expect(currentStep('sending')).toBe('sending')
    expect(currentStep('completed')).toBe('done')
  })
})

describe('requests', () => {
  it('getProgress gets the progress', async () => {
    vi.mocked(apiFetch).mockResolvedValue(progress())
    expect((await getProgress(12)).campaign_id).toBe(12)
    expect(apiFetch).toHaveBeenCalledWith('/campaigns/12/progress')
  })

  it('listSendJobs sends only the set filters', async () => {
    vi.mocked(apiFetch).mockResolvedValue({ send_jobs: [], next_after: null })
    await listSendJobs(12, { status: 'failed', channel: 'sms', after: 40, limit: 50 })
    expect(apiFetch).toHaveBeenCalledWith('/campaigns/12/send-jobs', { query: { status: 'failed', channel: 'sms', after: 40, limit: 50 } })
    await listSendJobs(12, {})
    expect(apiFetch).toHaveBeenLastCalledWith('/campaigns/12/send-jobs', { query: {} })
  })

  it('retryCampaign posts without a body', async () => {
    vi.mocked(apiFetch).mockResolvedValue({ retried: 3, progress: progress('sending') })
    expect((await retryCampaign(12)).retried).toBe(3)
    expect(apiFetch).toHaveBeenCalledWith('/campaigns/12/retry', { method: 'POST' })
  })
})

describe('watchProgress', () => {
  it('opens the stream and calls on for each progress event', () => {
    const on = vi.fn()
    watchProgress(12, on)
    const es = FakeEventSource.instances[0]!
    expect(es.url).toBe(`${API_BASE_URL}/campaigns/12/progress/stream`)
    es.emit('progress', progress('sending', 1))
    es.emit('progress', progress('sending', 2))
    expect(on).toHaveBeenCalledTimes(2)
    expect(on.mock.calls[1]![0].channels.sms.sent).toBe(2)
  })

  it('closes the stream at done and does not poll afterwards', () => {
    vi.useFakeTimers()
    const on = vi.fn()
    watchProgress(12, on)
    const es = FakeEventSource.instances[0]!
    es.emit('progress', progress('completed', 10))
    es.emit('done', {})
    expect(on).toHaveBeenCalledTimes(1)
    expect(on).toHaveBeenCalledWith(expect.objectContaining({ status: 'completed' }))
    expect(es.closed).toBe(true)
    es.emit('error')
    vi.advanceTimersByTime(5000)
    expect(apiFetch).not.toHaveBeenCalled()
  })

  it('polls every 2 seconds after a stream error', async () => {
    vi.useFakeTimers()
    vi.mocked(apiFetch).mockResolvedValue(progress('sending', 5))
    const on = vi.fn()
    watchProgress(12, on)
    const es = FakeEventSource.instances[0]!
    es.emit('error')
    expect(es.closed).toBe(true)
    await vi.advanceTimersByTimeAsync(2000)
    expect(apiFetch).toHaveBeenCalledTimes(1)
    expect(on).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(2000)
    expect(apiFetch).toHaveBeenCalledTimes(2)
  })

  it('skips a tick while a poll is in flight', async () => {
    vi.useFakeTimers()
    let resolve!: (p: Progress) => void
    vi.mocked(apiFetch).mockImplementation(() => new Promise((r) => {
      resolve = r as (p: Progress) => void
    }))
    const on = vi.fn()
    watchProgress(12, on)
    FakeEventSource.instances[0]!.emit('error')
    await vi.advanceTimersByTimeAsync(6000)
    expect(apiFetch).toHaveBeenCalledTimes(1)
    resolve(progress())
    await vi.advanceTimersByTimeAsync(2000)
    expect(apiFetch).toHaveBeenCalledTimes(2)
    expect(on).toHaveBeenCalledTimes(1)
  })

  it('keeps polling after a failed poll', async () => {
    vi.useFakeTimers()
    vi.mocked(apiFetch).mockRejectedValueOnce(new Error('net')).mockResolvedValue(progress())
    const on = vi.fn()
    watchProgress(12, on)
    FakeEventSource.instances[0]!.emit('error')
    await vi.advanceTimersByTimeAsync(4000)
    expect(apiFetch).toHaveBeenCalledTimes(2)
    expect(on).toHaveBeenCalledTimes(1)
  })

  it('stop closes the stream and the timer', async () => {
    vi.useFakeTimers()
    vi.mocked(apiFetch).mockResolvedValue(progress())
    const on = vi.fn()
    const stop = watchProgress(12, on)
    const es = FakeEventSource.instances[0]!
    stop()
    expect(es.closed).toBe(true)
    es.emit('progress', progress())
    expect(on).not.toHaveBeenCalled()

    const stop2 = watchProgress(12, on)
    FakeEventSource.instances[1]!.emit('error')
    stop2()
    await vi.advanceTimersByTimeAsync(6000)
    expect(apiFetch).not.toHaveBeenCalled()
  })

  it('ignores a poll that returns after stop', async () => {
    vi.useFakeTimers()
    let resolve!: (p: Progress) => void
    vi.mocked(apiFetch).mockImplementation(() => new Promise((r) => {
      resolve = r as (p: Progress) => void
    }))
    const on = vi.fn()
    const stop = watchProgress(12, on)
    FakeEventSource.instances[0]!.emit('error')
    await vi.advanceTimersByTimeAsync(2000)
    stop()
    resolve(progress())
    await vi.advanceTimersByTimeAsync(0)
    expect(on).not.toHaveBeenCalled()
  })
})
