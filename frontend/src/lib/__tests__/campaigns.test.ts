import { FetchError } from 'ofetch'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import { apiFetch } from '@/lib/api-client'
import { approveCampaign, CampaignRequestError, getCampaign, getSampleNames, listCampaigns, saveContents } from '@/lib/campaigns'

vi.mock('@/lib/api-client', () => ({ apiFetch: vi.fn() }))

const contents = [{ language: 'tr', sms: 's', email_subject: 'e', email_body: 'b' }]

describe('campaigns client', () => {
  beforeEach(() => {
    vi.mocked(apiFetch).mockReset()
  })

  it('listCampaigns unwraps the list', async () => {
    vi.mocked(apiFetch).mockResolvedValue({ campaigns: [{ id: 1 }] })
    expect(await listCampaigns()).toEqual([{ id: 1 }])
    expect(apiFetch).toHaveBeenCalledWith('/campaigns')
  })

  it('getCampaign gets by id', async () => {
    vi.mocked(apiFetch).mockResolvedValue({ id: 3 })
    expect(await getCampaign(3)).toEqual({ id: 3 })
    expect(apiFetch).toHaveBeenCalledWith('/campaigns/3')
  })

  it('getSampleNames maps languages to names', async () => {
    vi.mocked(apiFetch).mockResolvedValue({ languages: [{ code: 'tr', first_names: ['Ayşe'] }] })
    expect(await getSampleNames(3)).toEqual({ tr: ['Ayşe'] })
    expect(apiFetch).toHaveBeenCalledWith('/campaigns/3/sample')
  })

  it('saveContents sends PUT', async () => {
    vi.mocked(apiFetch).mockResolvedValue({ id: 3 })
    await saveContents(3, contents)
    expect(apiFetch).toHaveBeenCalledWith('/campaigns/3/contents', { method: 'PUT', body: { contents } })
  })

  it('approveCampaign sends POST', async () => {
    vi.mocked(apiFetch).mockResolvedValue({ id: 3, status: 'approved' })
    expect((await approveCampaign(3, contents)).status).toBe('approved')
    expect(apiFetch).toHaveBeenCalledWith('/campaigns/3/approve', { method: 'POST', body: { contents } })
  })

  it('rejects with CampaignRequestError and details on 422', async () => {
    const err = new FetchError('422')
    err.data = { error: { code: 'content_invalid', message: 'Bad.', details: [{ path: 'contents[0].sms', message: 'too long' }] } }
    err.statusCode = 422
    vi.mocked(apiFetch).mockImplementation(() => Promise.reject(err))
    const got = await approveCampaign(3, contents).catch(e => e)
    expect(got).toBeInstanceOf(CampaignRequestError)
    expect(got.code).toBe('content_invalid')
    expect(got.message).toBe('Bad.')
    expect(got.status).toBe(422)
    expect(got.details).toEqual([{ path: 'contents[0].sms', message: 'too long' }])
  })

  it('falls back on a network error', async () => {
    vi.mocked(apiFetch).mockImplementation(() => Promise.reject(new Error('net')))
    const got = await getCampaign(1).catch(e => e)
    expect(got).toBeInstanceOf(CampaignRequestError)
    expect(got.message).toBe('Something went wrong. Try again.')
    expect(got.details).toEqual([])
  })
})
