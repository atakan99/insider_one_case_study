import { FetchError } from 'ofetch'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { apiFetch } from '@/lib/api-client'
import { apiErrorMessage, previewSegment } from '@/lib/segments'

vi.mock('@/lib/api-client', () => ({ apiFetch: vi.fn() }))

describe('segments client', () => {
  beforeEach(() => {
    vi.mocked(apiFetch).mockReset()
  })

  it('previewSegment uses a 60s timeout', async () => {
    vi.mocked(apiFetch).mockResolvedValue({ description: 'x' })
    await previewSegment('hello')
    expect(apiFetch).toHaveBeenCalledWith('/segments/preview', { method: 'POST', body: { brief: 'hello' }, timeout: 60000 })
  })

  it('apiErrorMessage reads error.message from a FetchError body', () => {
    const err = new FetchError('boom')
    err.data = { error: { code: 'llm_unavailable', message: 'The model is down.' } }
    expect(apiErrorMessage(err)).toBe('The model is down.')
  })

  it('apiErrorMessage falls back for an Error without data', () => {
    expect(apiErrorMessage(new Error('x'))).toBe('Something went wrong. Try again.')
    expect(apiErrorMessage(null)).toBe('Something went wrong. Try again.')
  })
})
