import { afterEach, describe, expect, it, vi } from 'vitest'

import type { ContentEvent } from '@/lib/content'

import { API_BASE_URL } from '@/constants/app-config'
import { ContentRequestError, generateContent, parseSSE } from '@/lib/content'

const enc = new TextEncoder()

function sseResponse(chunks: string[], status = 200): Response {
  const stream = new ReadableStream<Uint8Array>({
    start(c) {
      for (const ch of chunks) c.enqueue(enc.encode(ch))
      c.close()
    },
  })
  return new Response(stream, { status, headers: { 'Content-Type': 'text/event-stream' } })
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('parseSSE', () => {
  it('splits events across chunks', () => {
    const first = parseSSE('event: draft\ndata: {"language":"tr","con', '')
    expect(first.events).toEqual([])
    const second = parseSSE('tent":{"sms":"a","email_subject":"","email_body":""}}\n\nevent: done\ndata: {}\n\n', first.carry)
    expect(second.events.map(e => e.event)).toEqual(['draft', 'done'])
    expect(second.carry).toBe('')
  })

  it('handles CRLF, comments and unknown events', () => {
    const r = parseSSE(': keepalive\r\n\r\nevent: x\r\ndata: {}\r\n\r\nevent: retry\r\ndata: {"language":"tr"}\r\n\r\n', '')
    expect(r.events).toEqual([{ event: 'retry', data: { language: 'tr' } }])
  })

  it('handles a CRLF split between chunks', () => {
    const a = parseSSE('event: done\r\ndata: {}\r\n\r', '')
    const b = parseSSE('\n', a.carry)
    expect(a.events).toEqual([])
    expect(b.events).toEqual([{ event: 'done', data: {} }])
  })
})

describe('generateContent', () => {
  it('posts brief and rule', async () => {
    const fetchMock = vi.fn().mockResolvedValue(sseResponse(['event: done\ndata: {}\n\n']))
    vi.stubGlobal('fetch', fetchMock)
    const events: ContentEvent[] = []
    await generateContent('hi', { op: 'and' }, e => events.push(e), new AbortController().signal)
    const [url, init] = fetchMock.mock.calls[0]
    expect(url).toBe(`${API_BASE_URL}/content/generate`)
    expect(init.method).toBe('POST')
    expect(JSON.parse(init.body)).toEqual({ brief: 'hi', rule: { op: 'and' } })
    expect(events).toEqual([{ event: 'done', data: {} }])
  })

  it('throws the API message on 400', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(
      JSON.stringify({ error: { code: 'invalid_brief', message: 'Brief is too short.' } }),
      { status: 400, headers: { 'Content-Type': 'application/json' } },
    )))
    const err = await generateContent('x', {}, () => {}, new AbortController().signal).catch(e => e)
    expect(err).toBeInstanceOf(ContentRequestError)
    expect(err.code).toBe('invalid_brief')
    expect(err.message).toBe('Brief is too short.')
  })

  it('falls back when the error body is not the error shape', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('oops', { status: 502 })))
    const err = await generateContent('x', {}, () => {}, new AbortController().signal).catch(e => e)
    expect(err.message).toBe('Something went wrong. Try again.')
  })

  it('rejects with stream_closed when the stream ends without done', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(sseResponse(['event: audience\ndata: {"description":"d","languages":[]}\n\n'])))
    const events: ContentEvent[] = []
    const err = await generateContent('x', {}, e => events.push(e), new AbortController().signal).catch(e => e)
    expect(err).toBeInstanceOf(ContentRequestError)
    expect(err.code).toBe('stream_closed')
    expect(err.message).toBe('The connection closed early. Try again.')
    expect(events).toHaveLength(1)
  })

  it('stops reading after abort and resolves quietly', async () => {
    const ctrl = new AbortController()
    let push!: (s: string) => void
    const stream = new ReadableStream<Uint8Array>({
      start(c) {
        push = s => c.enqueue(enc.encode(s))
      },
    })
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(stream, { status: 200 })))
    const events: ContentEvent[] = []
    const p = generateContent('x', {}, e => events.push(e), ctrl.signal)
    await vi.waitFor(() => expect(push).toBeTypeOf('function'))
    push('event: retry\ndata: {"language":"tr"}\n\n')
    await vi.waitFor(() => expect(events).toHaveLength(1))
    ctrl.abort()
    push('event: retry\ndata: {"language":"en"}\n\n')
    await expect(p).resolves.toBeUndefined()
    expect(events).toHaveLength(1)
  })
})
