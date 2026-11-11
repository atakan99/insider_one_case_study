import type { CampaignStatus } from '@/lib/campaigns'
import type { Counts } from '@/lib/segments'

import { API_BASE_URL } from '@/constants/app-config'

export interface Content { sms: string, email_subject: string, email_body: string }
export interface ContentLimits { first_name_runes: number, sms_runes: number, email_subject_runes: number, email_body_runes: number }
// 'restart': the model began the text again (not the repair round, no errors).
// 'invalid': the text failed the check, and the repair round starts.
export type RetryReason = 'restart' | 'invalid'
export type ContentEvent
  = | { event: 'audience', data: { description: string, languages: (Counts & { code: string })[] } }
    | { event: 'draft', data: { language: string, content: Content } }
    | { event: 'retry', data: { language: string, reason?: RetryReason, errors?: { path: string, message: string }[] } }
    | { event: 'final', data: { language: string, content: Content, limits: ContentLimits } }
    | { event: 'error', data: { language?: string, code: string, message: string } }
    | { event: 'campaign', data: { id: number, status: CampaignStatus } }
    | { event: 'done', data: Record<string, never> }

const FALLBACK_MESSAGE = 'Something went wrong. Try again.'
const KNOWN_EVENTS = new Set(['audience', 'draft', 'retry', 'final', 'error', 'campaign', 'done'])

export class ContentRequestError extends Error {
  code: string

  constructor(code: string, message: string) {
    super(message)
    this.name = 'ContentRequestError'
    this.code = code
  }
}

function parseBlock(block: string): ContentEvent | null {
  let name = ''
  const dataLines: string[] = []
  for (const line of block.split('\n')) {
    if (line === '' || line.startsWith(':'))
      continue
    const idx = line.indexOf(':')
    const field = idx === -1 ? line : line.slice(0, idx)
    let value = idx === -1 ? '' : line.slice(idx + 1)
    if (value.startsWith(' '))
      value = value.slice(1)
    if (field === 'event')
      name = value
    else if (field === 'data')
      dataLines.push(value)
  }
  if (!KNOWN_EVENTS.has(name) || dataLines.length === 0)
    return null
  try {
    return { event: name, data: JSON.parse(dataLines.join('\n')) } as ContentEvent
  }
  catch {
    return null
  }
}

// Returns the complete events in `carry + chunk` and the unfinished rest.
export function parseSSE(chunk: string, carry: string): { events: ContentEvent[], carry: string } {
  const text = (carry + chunk).replace(/\r\n/g, '\n')
  const parts = text.split('\n\n')
  const rest = parts.pop() ?? ''
  const events: ContentEvent[] = []
  for (const part of parts) {
    const ev = parseBlock(part)
    if (ev)
      events.push(ev)
  }
  return { events, carry: rest }
}

async function requestError(res: Response): Promise<ContentRequestError> {
  try {
    const body = await res.json() as { error?: { code?: unknown, message?: unknown } }
    const code = typeof body?.error?.code === 'string' ? body.error.code : 'request_failed'
    const message = typeof body?.error?.message === 'string' && body.error.message !== '' ? body.error.message : FALLBACK_MESSAGE
    return new ContentRequestError(code, message)
  }
  catch {
    return new ContentRequestError('request_failed', FALLBACK_MESSAGE)
  }
}

function isAbort(err: unknown): boolean {
  return (err as { name?: string } | null)?.name === 'AbortError'
}

// After an abort the promise resolves quietly and onEvent is not called again.
export async function generateContent(
  brief: string,
  rule: unknown,
  onEvent: (e: ContentEvent) => void,
  signal: AbortSignal,
  campaignId?: number,
): Promise<void> {
  try {
    const res = await fetch(`${API_BASE_URL}/content/generate`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Accept': 'text/event-stream' },
      body: JSON.stringify(campaignId === undefined ? { brief, rule } : { brief, rule, campaign_id: campaignId }),
      signal,
    })
    if (!res.ok)
      throw await requestError(res)
    if (!res.body)
      throw new ContentRequestError('stream_closed', 'The connection closed early. Try again.')

    const reader = res.body.getReader()
    const decoder = new TextDecoder()
    let carry = ''
    let done = false
    for (;;) {
      const { value, done: finished } = await reader.read()
      if (signal.aborted)
        return
      const parsed = parseSSE(finished ? `${decoder.decode()}\n\n` : decoder.decode(value, { stream: true }), carry)
      carry = parsed.carry
      for (const ev of parsed.events) {
        if (signal.aborted)
          return
        onEvent(ev)
        if (ev.event === 'done')
          done = true
      }
      if (done)
        return
      if (finished)
        break
    }
    throw new ContentRequestError('stream_closed', 'The connection closed early. Try again.')
  }
  catch (err) {
    if (signal.aborted || isAbort(err))
      return
    if (err instanceof ContentRequestError)
      throw err
    throw new ContentRequestError('stream_closed', 'The connection closed early. Try again.')
  }
}
