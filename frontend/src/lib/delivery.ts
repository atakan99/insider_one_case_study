import type { CampaignStatus } from '@/lib/campaigns'

import { API_BASE_URL } from '@/constants/app-config'
import { apiFetch } from '@/lib/api-client'
import { CampaignRequestError } from '@/lib/campaigns'

export type DeliveryStatus = 'queued' | 'sending' | 'retrying' | 'sent' | 'failed' | 'unknown' | 'skipped'
export type DeliveryChannel = 'sms' | 'email'
export interface ChannelProgress { total: number, queued: number, sending: number, retrying: number, sent: number, failed: number, unknown: number, skipped: number, retryable_failed: number }
export interface Progress { campaign_id: number, status: CampaignStatus, channels: { sms: ChannelProgress, email: ChannelProgress } }
export interface SendJob { id: number, user_id: number, channel: DeliveryChannel, language: string, status: DeliveryStatus, attempts: number, last_error: string, updated_at: string }
export interface SendJobQuery { status?: DeliveryStatus, channel?: DeliveryChannel, after?: number, limit?: number }
export type Step = 'approved' | 'queued' | 'sending' | 'done'

const POLL_MS = 2000
const FALLBACK_MESSAGE = 'Something went wrong. Try again.'

async function call<T>(p: () => Promise<T>): Promise<T> {
  try {
    return await p()
  }
  catch (err) {
    const e = err as { data?: { error?: { code?: unknown, message?: unknown } }, statusCode?: unknown } | null
    const body = e?.data?.error
    throw new CampaignRequestError(
      typeof body?.code === 'string' ? body.code : 'request_failed',
      typeof body?.message === 'string' && body.message !== '' ? body.message : FALLBACK_MESSAGE,
      typeof e?.statusCode === 'number' ? e.statusCode : 0,
      [],
    )
  }
}

export function getProgress(id: number): Promise<Progress> {
  return call(() => apiFetch<Progress>(`/campaigns/${id}/progress`))
}

export function listSendJobs(id: number, q: SendJobQuery): Promise<{ send_jobs: SendJob[], next_after: number | null }> {
  const query: Record<string, string | number> = {}
  if (q.status)
    query.status = q.status
  if (q.channel)
    query.channel = q.channel
  if (q.after !== undefined)
    query.after = q.after
  if (q.limit !== undefined)
    query.limit = q.limit
  return call(() => apiFetch<{ send_jobs: SendJob[], next_after: number | null }>(`/campaigns/${id}/send-jobs`, { query }))
}

export function retryCampaign(id: number): Promise<{ retried: number, progress: Progress }> {
  return call(() => apiFetch<{ retried: number, progress: Progress }>(`/campaigns/${id}/retry`, { method: 'POST' }))
}

export function currentStep(status: CampaignStatus): Step {
  switch (status) {
    case 'queued': return 'queued'
    case 'sending': return 'sending'
    case 'completed': return 'done'
    default: return 'approved'
  }
}

function parse(e: Event): Progress | null {
  try {
    const p = JSON.parse((e as MessageEvent).data) as Progress
    return p && typeof p === 'object' && p.channels ? p : null
  }
  catch {
    return null
  }
}

// Reads the progress stream. If the stream fails, it asks for the progress every 2 s.
export function watchProgress(id: number, on: (p: Progress) => void): () => void {
  let stopped = false
  let inFlight = false
  let timer: ReturnType<typeof setInterval> | undefined
  const es = new EventSource(`${API_BASE_URL}/campaigns/${id}/progress/stream`)

  function stop() {
    stopped = true
    es.close()
    if (timer !== undefined) {
      clearInterval(timer)
      timer = undefined
    }
  }

  function emit(p: Progress | null) {
    if (!stopped && p)
      on(p)
  }

  es.addEventListener('progress', e => emit(parse(e)))
  es.addEventListener('done', (e) => {
    emit(parse(e))
    stop()
  })
  es.addEventListener('error', () => {
    if (stopped)
      return
    es.close()
    timer ??= setInterval(async () => {
      if (inFlight)
        return
      inFlight = true
      try {
        const p = await getProgress(id)
        emit(p)
        if (p.status === 'completed')
          stop()
      }
      catch {
        // Try again at the next tick.
      }
      finally {
        inFlight = false
      }
    }, POLL_MS)
  })
  return stop
}
