import type { CampaignStatus } from '@/lib/campaigns'
import type { DeliveryChannel, DeliveryStatus } from '@/lib/delivery'

import { apiFetch } from '@/lib/api-client'
import { CampaignRequestError } from '@/lib/campaigns'

export interface ChannelTotals { sent: number, failed: number, unknown: number, skipped: number, in_progress: number }
export interface DashboardStats {
  generated_at: string
  totals: Record<DeliveryChannel, ChannelTotals>
  throughput: { bucket_seconds: number, start: string, sms: number[], email: number[] }
  failures: { status: DeliveryStatus, reason: string, sms: number, email: number }[]
  mean_attempts: Record<DeliveryChannel, number | null>
  campaigns: { id: number, description: string, status: CampaignStatus, total: number, sent: number, failed: number, unknown: number, duration_ms: number | null }[]
}
export interface DashboardHealth {
  services: Record<'db' | 'redis' | 'worker' | 'provider', 'ok' | 'down'>
  streams: { channel: DeliveryChannel, length: number | null, pending: number | null, in_progress: number }[]
}

export const DASHBOARD_POLL_MS = 5000

const FALLBACK_MESSAGE = 'Something went wrong. Try again.'
const DASH = '–'

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

export function getDashboardStats(): Promise<DashboardStats> {
  return call(() => apiFetch<DashboardStats>('/dashboard/stats'))
}

export function getDashboardHealth(): Promise<DashboardHealth> {
  return call(() => apiFetch<DashboardHealth>('/dashboard/health'))
}

export function percent(part: number, total: number): string {
  if (total <= 0)
    return DASH
  return `${(part / total * 100).toFixed(1)}%`
}

/** Share of finished jobs (sent, failed, unknown) that were sent. */
export function successRate(t: ChannelTotals): string {
  return percent(t.sent, t.sent + t.failed + t.unknown)
}

export function formatDuration(ms: number | null): string {
  if (ms === null)
    return DASH
  if (ms < 60000)
    return `${(ms / 1000).toFixed(1)} s`
  const total = Math.round(ms / 1000)
  return `${Math.floor(total / 60)} min ${total % 60} s`
}
