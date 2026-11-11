// @vitest-environment happy-dom

import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { toast } from 'vue-sonner'

import type { Campaign } from '@/lib/campaigns'

import { approveCampaign, CampaignRequestError, getCampaign, getSampleNames, saveContents } from '@/lib/campaigns'

import ReviewDialog from '../review-dialog.vue'

vi.mock('@/lib/campaigns', async (orig) => {
  const real = await orig<typeof import('@/lib/campaigns')>()
  return {
    ...real,
    getCampaign: vi.fn(),
    getSampleNames: vi.fn(),
    saveContents: vi.fn(),
    approveCampaign: vi.fn(),
  }
})
vi.mock('vue-sonner', () => ({ toast: { success: vi.fn(), error: vi.fn() } }))

function campaign(over: Partial<Campaign> = {}): Campaign {
  return {
    id: 12,
    status: 'draft',
    brief: 'b',
    description: 'Istanbul users',
    skipped_languages: [],
    created_at: '2026-11-10T12:00:00Z',
    updated_at: '2026-11-10T12:00:00Z',
    approved_at: null,
    contents: [
      { language: 'en', sms: 'Hi {{first_name}}', email_subject: 'Sale {{first_name}}', email_body: 'Hello {{first_name}}', sms_count: 20, email_count: 30, first_name_runes: 9 },
      { language: 'tr', sms: 'Merhaba {{first_name}}', email_subject: 'İndirim', email_body: 'Merhaba {{first_name}},\nSepet', sms_count: 980, email_count: 1212, first_name_runes: 14 },
    ],
    ...over,
  }
}

const getC = vi.mocked(getCampaign)
const getS = vi.mocked(getSampleNames)
const save = vi.mocked(saveContents)
const approve = vi.mocked(approveCampaign)

async function make(c: Campaign = campaign()) {
  getC.mockResolvedValue(c)
  const w = mount(ReviewDialog, { props: { 'campaignId': c.id, 'open': true, 'onUpdate:open': (v: boolean) => w.setProps({ open: v }) }, attachTo: document.body })
  await flushPromises()
  return w
}

const $ = (sel: string) => document.body.querySelector<HTMLElement>(sel)
const $$ = (sel: string) => [...document.body.querySelectorAll<HTMLElement>(sel)]
const text = (sel: string) => $(sel)?.textContent?.trim() ?? ''

async function pressTab(sel: string) {
  const el = $(sel)!
  el.dispatchEvent(new MouseEvent('mousedown', { bubbles: true, button: 0 }))
  el.dispatchEvent(new MouseEvent('click', { bubbles: true, button: 0 }))
  await flushPromises()
}

async function type(sel: string, value: string) {
  const el = $(sel) as HTMLInputElement | HTMLTextAreaElement
  el.value = value
  el.dispatchEvent(new Event('input', { bubbles: true }))
  await flushPromises()
}

async function click(sel: string) {
  $(sel)!.click()
  await flushPromises()
}

beforeEach(() => {
  getC.mockReset()
  getS.mockReset()
  save.mockReset()
  approve.mockReset()
  vi.mocked(toast.success).mockReset()
  getS.mockResolvedValue({ tr: ['Abdurrahmangazi', 'Ayşe', 'Mehmet'], en: ['Christopher', 'Ann'] })
})

afterEach(() => {
  document.body.innerHTML = ''
})

describe('review dialog', () => {
  it('shows tabs for each channel and language', async () => {
    await make()
    expect($$('[data-test="channel-tab"]').map(e => e.dataset.value)).toEqual(['sms', 'email'])
    expect($$('[data-test="lang-tab"]').map(e => e.textContent?.trim())).toEqual(['English', 'Turkish'])
    expect(text('[data-test="channel-tab"][data-value="sms"]')).toContain('20')
    expect($('[data-test="sms-input"]')).not.toBeNull()
    expect($('[data-test="sms-preview"]')).not.toBeNull()
    await pressTab('[data-test="channel-tab"][data-value="email"]')
    expect($('[data-test="subject-input"]')).not.toBeNull()
    expect($('[data-test="body-input"]')).not.toBeNull()
    expect($('[data-test="email-preview"]')).not.toBeNull()
    expect(text('[data-test="email-preview"]')).toContain('Your brand')
  })

  it('counter at 160 and 161', async () => {
    await make()
    await pressTab('[data-test="lang-tab"][data-value="tr"]')
    await type('[data-test="sms-input"]', `{{first_name}}${'a'.repeat(146)}`)
    expect(text('[data-test="sms-counter"]')).toBe('160 / 160')
    expect($('[data-test="approve"]')!.hasAttribute('disabled')).toBe(false)
    await type('[data-test="sms-input"]', `{{first_name}}${'a'.repeat(147)}`)
    expect(text('[data-test="sms-counter"]')).toBe('161 / 160')
    expect($('[data-test="sms-counter"]')!.className).toContain('text-destructive')
    expect($('[data-test="approve"]')!.hasAttribute('disabled')).toBe(true)
    expect($('[data-test="save"]')!.hasAttribute('disabled')).toBe(true)
  })

  it('counter warns above 90%', async () => {
    await make()
    await pressTab('[data-test="lang-tab"][data-value="tr"]')
    await type('[data-test="sms-input"]', `{{first_name}}${'a'.repeat(132)}`) // 146
    expect($('[data-test="sms-counter"]')!.className).toContain('text-warning')
  })

  it('shows field errors and a tab mark', async () => {
    await make()
    await pressTab('[data-test="lang-tab"][data-value="tr"]')
    await type('[data-test="sms-input"]', 'Hi {{last_name}}')
    expect(text('[data-test="sms-error"]')).toBe('Only {{first_name}} is allowed.')
    expect($('[data-test="lang-tab"][data-value="tr"] [data-test="tab-error"]')).not.toBeNull()
    expect($('[data-test="lang-tab"][data-value="en"] [data-test="tab-error"]')).toBeNull()
    expect($('[data-test="approve"]')!.hasAttribute('disabled')).toBe(true)
    await type('[data-test="sms-input"]', '  ')
    expect(text('[data-test="sms-error"]')).toBe('Write some text.')
    await type('[data-test="sms-input"]', 'ok')
    await pressTab('[data-test="channel-tab"][data-value="email"]')
    // A single-line input drops line breaks, so the browser cannot send one.
    await type('[data-test="subject-input"]', 'a\nb')
    expect(($('[data-test="subject-input"]') as HTMLInputElement).value).not.toContain('\n')
    await type('[data-test="subject-input"]', `{{first_name}}${'b'.repeat(67)}`)
    expect(text('[data-test="subject-error"]')).toBe('Too long after the name is filled in.')
  })

  it('preview uses the longest sample first', async () => {
    await make()
    await pressTab('[data-test="lang-tab"][data-value="tr"]')
    expect(text('[data-test="sms-preview"]')).toBe('Merhaba Abdurrahmangazi')
    expect(text('[data-test="name-select"]')).toContain('Abdurrahmangazi (longest)')
  })

  it('renders html as text', async () => {
    await make()
    await type('[data-test="sms-input"]', '<b>x</b> {{first_name}}')
    expect($('[data-test="sms-preview"] b')).toBeNull()
    expect(text('[data-test="sms-preview"]')).toBe('<b>x</b> Christopher')
  })

  it('falls back to Alex when the sample fails', async () => {
    getS.mockImplementation(() => Promise.reject(new CampaignRequestError('db_unavailable', 'x', 503, [])))
    await make()
    expect(text('[data-test="sms-preview"]')).toBe('Hi Alex')
    expect($('[data-test="sample-note"]')).not.toBeNull()
    expect($('[data-test="approve"]')!.hasAttribute('disabled')).toBe(false)
  })

  it('names skipped languages and totals in confirm', async () => {
    await make(campaign({ skipped_languages: ['de'] }))
    expect(text('[data-test="footer-note"]')).toContain('DE skipped: content failed.')
    await click('[data-test="approve"]')
    const confirm = text('[data-test="confirm"]')
    expect(confirm).toContain('Send this campaign?')
    expect(confirm).toMatch(/SMS\s*1,000/)
    expect(confirm).toMatch(/Email\s*1,242/)
    expect(confirm).toContain('DE skipped')
  })

  it('save draft sends one PUT and toasts', async () => {
    save.mockResolvedValue(campaign())
    await make()
    await type('[data-test="sms-input"]', 'Hey {{first_name}}')
    await click('[data-test="save"]')
    expect(save).toHaveBeenCalledTimes(1)
    const [id, contents] = save.mock.calls[0]!
    expect(id).toBe(12)
    expect(contents).toEqual([
      { language: 'en', sms: 'Hey {{first_name}}', email_subject: 'Sale {{first_name}}', email_body: 'Hello {{first_name}}' },
      { language: 'tr', sms: 'Merhaba {{first_name}}', email_subject: 'İndirim', email_body: 'Merhaba {{first_name}},\nSepet' },
    ])
    expect(toast.success).toHaveBeenCalledWith('Draft saved.')
  })

  it('keeps the text typed while a save is in progress', async () => {
    let finish!: (c: Campaign) => void
    save.mockReturnValue(new Promise<Campaign>((resolve) => {
      finish = resolve
    }))
    await make()
    await type('[data-test="sms-input"]', 'Hey {{first_name}}')
    await click('[data-test="save"]')
    await type('[data-test="sms-input"]', 'Hey {{first_name}}, one more thing')
    const c = campaign()
    c.contents[0]!.sms = 'Hey {{first_name}} (saved)'
    c.contents[0]!.email_body = 'Body from the server'
    finish(c)
    await flushPromises()
    const sms = () => ($('[data-test="sms-input"]') as HTMLTextAreaElement).value
    expect(sms()).toBe('Hey {{first_name}}, one more thing')
    expect(toast.success).toHaveBeenCalledWith('Draft saved.')
    // A field with no newer edit takes the saved value.
    await pressTab('[data-test="channel-tab"][data-value="email"]')
    expect(($('[data-test="body-input"]') as HTMLTextAreaElement).value).toBe('Body from the server')
    await pressTab('[data-test="channel-tab"][data-value="sms"]')
    // The newer text is not saved, so a close asks.
    await click('[data-test="cancel"]')
    expect($('[data-test="discard"]')).not.toBeNull()
  })

  it('cancel asks only when dirty', async () => {
    const w = await make()
    await click('[data-test="cancel"]')
    expect(w.emitted('update:open')?.at(-1)).toEqual([false])
    const w2 = await make()
    await type('[data-test="sms-input"]', 'changed')
    await click('[data-test="cancel"]')
    expect(w2.emitted('update:open')).toBeUndefined()
    expect(text('[data-test="discard"]')).toContain('Discard your changes?')
    await click('[data-test="discard-confirm"]')
    expect(w2.emitted('update:open')?.at(-1)).toEqual([false])
  })

  it('asks before closing with Escape', async () => {
    const w = await make()
    await type('[data-test="sms-input"]', 'changed')
    document.activeElement?.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape', bubbles: true }))
    await flushPromises()
    expect(w.emitted('update:open')).toBeUndefined()
    expect(text('[data-test="discard"]')).toContain('Discard your changes?')
  })

  it('approve sends once on double click and emits approved', async () => {
    let resolve: (c: Campaign) => void = () => {}
    approve.mockImplementation(() => new Promise((r) => {
      resolve = r
    }))
    const w = await make()
    await click('[data-test="approve"]')
    $('[data-test="send-now"]')!.click()
    $('[data-test="send-now"]')!.click()
    await flushPromises()
    expect(approve).toHaveBeenCalledTimes(1)
    expect($('[data-test="send-now"]')!.hasAttribute('disabled')).toBe(true)
    expect($('[data-test="confirm-back"]')!.hasAttribute('disabled')).toBe(true)
    resolve(campaign({ status: 'approved' }))
    await flushPromises()
    expect(w.emitted('approved')?.[0]).toEqual([12])
  })

  it('back keeps the edits', async () => {
    await make()
    await type('[data-test="sms-input"]', 'Edited {{first_name}}')
    await click('[data-test="approve"]')
    await click('[data-test="confirm-back"]')
    expect(($('[data-test="sms-input"]') as HTMLTextAreaElement).value).toBe('Edited {{first_name}}')
  })

  it('maps 422 details to fields and banner', async () => {
    approve.mockImplementation(() => Promise.reject(new CampaignRequestError('content_invalid', 'Some messages are not valid.', 422, [
      { path: 'contents[1].sms', message: 'sms: 171 runes after rendering, limit 160' },
      { path: 'contents[0].language', message: 'no user of language "en" can get a message now' },
    ])))
    await make()
    await click('[data-test="approve"]')
    await click('[data-test="send-now"]')
    expect($('[data-test="confirm"]')).toBeNull()
    expect(text('[data-test="banner"]')).toContain('no user of language "en"')
    expect($('[data-test="lang-tab"][data-value="tr"] [data-test="tab-error"]')).not.toBeNull()
    await pressTab('[data-test="lang-tab"][data-value="tr"]')
    expect(text('[data-test="sms-error"]')).toContain('171 runes')
  })

  it('approved campaign is read-only', async () => {
    await make(campaign({ status: 'approved', approved_at: '2026-11-10T12:00:00Z' }))
    expect($('[data-test="sms-input"]')).toBeNull()
    expect($('[data-test="save"]')).toBeNull()
    expect($('[data-test="approve"]')).toBeNull()
    expect(text('[data-test="sms-preview"]')).toBe('Hi Christopher')
  })

  it('loads again on each open when nothing is unsaved', async () => {
    const w = await make()
    expect(($('[data-test="sms-input"]') as HTMLTextAreaElement).value).toBe('Hi {{first_name}}')
    await w.setProps({ open: false })
    await flushPromises()
    const regenerated = campaign()
    regenerated.contents[0] = { ...regenerated.contents[0]!, sms: 'New {{first_name}}' }
    getC.mockResolvedValue(regenerated)
    await w.setProps({ open: true })
    await flushPromises()
    expect(getC).toHaveBeenCalledTimes(2)
    expect(($('[data-test="sms-input"]') as HTMLTextAreaElement).value).toBe('New {{first_name}}')
  })

  it('ignores a late response for another campaign', async () => {
    let resolveA: (c: Campaign) => void = () => {}
    getC.mockImplementationOnce(() => new Promise((r) => {
      resolveA = r
    }))
    const w = mount(ReviewDialog, { props: { campaignId: 1, open: true }, attachTo: document.body })
    await flushPromises()
    getC.mockResolvedValueOnce(campaign({ id: 2 }))
    await w.setProps({ campaignId: 2 })
    await flushPromises()
    const a = campaign({ id: 1 })
    a.contents[0] = { ...a.contents[0]!, sms: 'From A' }
    resolveA(a)
    await flushPromises()
    expect(($('[data-test="sms-input"]') as HTMLTextAreaElement).value).toBe('Hi {{first_name}}')
  })

  it('takes the new longest name after a 422 and keeps the edits', async () => {
    approve.mockImplementation(() => Promise.reject(new CampaignRequestError('content_invalid', 'Some messages are not valid.', 422, [
      { path: 'contents[1].sms', message: 'sms: 161 runes after rendering, limit 160' },
    ])))
    await make()
    await pressTab('[data-test="lang-tab"][data-value="tr"]')
    await type('[data-test="sms-input"]', `{{first_name}}${'a'.repeat(140)}`) // 154 with 14
    const longer = campaign()
    longer.contents[1] = { ...longer.contents[1]!, first_name_runes: 21 }
    getC.mockResolvedValue(longer)
    await click('[data-test="approve"]')
    await click('[data-test="send-now"]')
    expect(($('[data-test="sms-input"]') as HTMLTextAreaElement).value).toBe(`{{first_name}}${'a'.repeat(140)}`)
    expect(text('[data-test="sms-counter"]')).toBe('161 / 160')
    await type('[data-test="sms-input"]', `{{first_name}}${'a'.repeat(141)}`)
    expect($('[data-test="approve"]')!.hasAttribute('disabled')).toBe(true)
  })
})
