import { apiFetch } from '@/lib/api-client'

export interface Counts { email: number, sms: number, reachable: number, matched: number, no_consent: number }
export interface SegmentPreview {
  rule: unknown
  description: string
  unmapped: string[]
  audience: Counts & { languages: (Counts & { code: string })[] }
}

const FALLBACK_MESSAGE = 'Something went wrong. Try again.'

// Two LLM calls can take about 40 s, so the default API timeout is too short.
const PREVIEW_TIMEOUT_MS = 60000

export async function previewSegment(brief: string): Promise<SegmentPreview> {
  return apiFetch<SegmentPreview>('/segments/preview', {
    method: 'POST',
    body: { brief },
    timeout: PREVIEW_TIMEOUT_MS,
  })
}

export function apiErrorMessage(err: unknown): string {
  const data = (err as { data?: { error?: { message?: unknown } } } | null)?.data
  const message = data?.error?.message
  return typeof message === 'string' && message !== '' ? message : FALLBACK_MESSAGE
}
