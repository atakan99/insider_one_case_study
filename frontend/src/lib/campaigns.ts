import type { Content } from '@/lib/content'

import { apiFetch } from '@/lib/api-client'

export type CampaignStatus = 'draft' | 'approved'
export interface LanguageContent extends Content { language: string, sms_count: number, email_count: number, first_name_runes: number }
export interface Campaign {
  id: number
  status: CampaignStatus
  brief: string
  description: string
  skipped_languages: string[]
  created_at: string
  updated_at: string
  approved_at: string | null
  contents: LanguageContent[]
}
export interface CampaignSummary { id: number, status: CampaignStatus, description: string, languages: string[], created_at: string, approved_at: string | null }
export interface FieldError { path: string, message: string }
export type ContentInput = Content & { language: string }

export const LANGUAGE_NAMES: Record<string, string> = { tr: 'Turkish', en: 'English' }

export function languageName(code: string): string {
  return LANGUAGE_NAMES[code] ?? code.toUpperCase()
}

const FALLBACK_MESSAGE = 'Something went wrong. Try again.'

export class CampaignRequestError extends Error {
  code: string
  details: FieldError[]
  status: number

  constructor(code: string, message: string, status: number, details: FieldError[]) {
    super(message)
    this.name = 'CampaignRequestError'
    this.code = code
    this.status = status
    this.details = details
  }
}

function toError(err: unknown): CampaignRequestError {
  const e = err as { data?: { error?: { code?: unknown, message?: unknown, details?: unknown } }, statusCode?: unknown } | null
  const body = e?.data?.error
  const code = typeof body?.code === 'string' ? body.code : 'request_failed'
  const message = typeof body?.message === 'string' && body.message !== '' ? body.message : FALLBACK_MESSAGE
  const status = typeof e?.statusCode === 'number' ? e.statusCode : 0
  const details = Array.isArray(body?.details)
    ? (body.details as unknown[]).filter((d): d is FieldError =>
        typeof (d as FieldError)?.path === 'string' && typeof (d as FieldError)?.message === 'string')
    : []
  return new CampaignRequestError(code, message, status, details)
}

async function call<T>(p: () => Promise<T>): Promise<T> {
  try {
    return await p()
  }
  catch (err) {
    throw toError(err)
  }
}

export async function listCampaigns(): Promise<CampaignSummary[]> {
  const res = await call(() => apiFetch<{ campaigns: CampaignSummary[] }>('/campaigns'))
  return res.campaigns
}

export function getCampaign(id: number): Promise<Campaign> {
  return call(() => apiFetch<Campaign>(`/campaigns/${id}`))
}

export async function getSampleNames(id: number): Promise<Record<string, string[]>> {
  const res = await call(() => apiFetch<{ languages: { code: string, first_names: string[] }[] }>(`/campaigns/${id}/sample`))
  return Object.fromEntries(res.languages.map(l => [l.code, l.first_names]))
}

export function saveContents(id: number, contents: ContentInput[]): Promise<Campaign> {
  return call(() => apiFetch<Campaign>(`/campaigns/${id}/contents`, { method: 'PUT', body: { contents } }))
}

export function approveCampaign(id: number, contents: ContentInput[]): Promise<Campaign> {
  return call(() => apiFetch<Campaign>(`/campaigns/${id}/approve`, { method: 'POST', body: { contents } }))
}
