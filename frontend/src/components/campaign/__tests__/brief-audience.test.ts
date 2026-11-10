// @vitest-environment happy-dom

import { VueQueryPlugin } from '@tanstack/vue-query'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import type { SegmentPreview } from '@/lib/segments'

import { previewSegment } from '@/lib/segments'

import BriefAudience from '../brief-audience.vue'

vi.mock('@/lib/segments', () => ({
  previewSegment: vi.fn(),
  apiErrorMessage: (e: unknown) => (e instanceof Error ? e.message : 'Something went wrong. Try again.'),
}))

function preview(over: Partial<SegmentPreview> = {}): SegmentPreview {
  return {
    rule: {},
    description: 'Turkish customers who bought recently',
    unmapped: ['Could not map "loyal"'],
    audience: {
      email: 290,
      sms: 205,
      reachable: 300,
      matched: 382,
      no_consent: 82,
      languages: [{ code: 'tr', email: 100, sms: 50, reachable: 110, matched: 120, no_consent: 10 }],
    },
    ...over,
  }
}

function make() {
  return mount(BriefAudience, { global: { plugins: [[VueQueryPlugin, {}]] } })
}

describe('brief audience', () => {
  beforeEach(() => {
    vi.mocked(previewSegment).mockReset()
  })

  it('disables the button for an empty brief and for 1001 characters', async () => {
    const w = make()
    const btn = () => w.find('button')
    expect(btn().attributes('disabled')).toBeDefined()
    await w.find('textarea').setValue('a'.repeat(1001))
    expect(w.text()).toContain('1001 / 1000')
    expect(btn().attributes('disabled')).toBeDefined()
    await w.find('textarea').setValue('a'.repeat(1000))
    expect(btn().attributes('disabled')).toBeUndefined()
  })

  it('shows progress while pending', async () => {
    vi.mocked(previewSegment).mockReturnValue(new Promise(() => {}))
    const w = make()
    await w.find('textarea').setValue('brief')
    await w.find('button').trigger('click')
    await flushPromises()
    expect(w.find('button').attributes('disabled')).toBeDefined()
    expect(w.text()).toContain('Generating segment…')
  })

  it('shows the audience on success', async () => {
    vi.mocked(previewSegment).mockResolvedValue(preview())
    const w = make()
    await w.find('textarea').setValue('brief')
    await w.find('button').trigger('click')
    await flushPromises()
    const t = w.text()
    expect(t).toContain('Turkish customers who bought recently')
    expect(t).toContain('290')
    expect(t).toContain('205')
    expect(t).toContain('excluded: no consent')
    expect(t).toContain('82')
    expect(t).toContain('Reachable')
    expect(t).toContain('300')
    expect(t).toContain('Not used from the brief')
    const cells = w.findAll('tbody tr').map(r => r.findAll('td').map(c => c.text()))
    expect(cells).toEqual([['tr', '100', '50', '10']])
    expect(t).toContain('Could not map "loyal"')
  })

  it('shows the error and keeps the brief', async () => {
    vi.mocked(previewSegment).mockRejectedValue(new Error('The model is down.'))
    const w = make()
    await w.find('textarea').setValue('my brief')
    await w.find('button').trigger('click')
    await flushPromises()
    expect(w.text()).toContain('The model is down.')
    expect((w.find('textarea').element as HTMLTextAreaElement).value).toBe('my brief')
  })

  it('renders unmapped notes as text', async () => {
    vi.mocked(previewSegment).mockResolvedValue(preview({ unmapped: ['<img src=x onerror=alert(1)>'] }))
    const w = make()
    await w.find('textarea').setValue('brief')
    await w.find('button').trigger('click')
    await flushPromises()
    expect(w.text()).toContain('<img src=x onerror=alert(1)>')
    expect(w.find('img').exists()).toBe(false)
  })

  it('replaces the first result with the second', async () => {
    vi.mocked(previewSegment).mockResolvedValueOnce(preview({ description: 'First result' }))
    vi.mocked(previewSegment).mockResolvedValueOnce(preview({ description: 'Second result' }))
    const w = make()
    await w.find('textarea').setValue('brief')
    await w.find('button').trigger('click')
    await flushPromises()
    await w.find('button').trigger('click')
    await flushPromises()
    expect(w.text()).toContain('Second result')
    expect(w.text()).not.toContain('First result')
  })
})
