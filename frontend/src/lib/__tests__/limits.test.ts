import { describe, expect, it } from 'vitest'

import { checkContent, LIMITS, renderedLength, runeLength } from '@/lib/limits'

const ok = { sms: 'Hi {{first_name}}', email_subject: 'Subject', email_body: 'Body' }

describe('limits', () => {
  it('has the case limits', () => {
    expect(LIMITS).toEqual({ sms: 160, email_subject: 80, email_body: 2000 })
  })

  it('counts code points, not UTF-16 units', () => {
    expect(runeLength('ş')).toBe(1)
    expect(runeLength('😀')).toBe(1)
  })

  it('renders with the longest name (slice 05 FR-9)', () => {
    const v = '{{first_name}}'
    expect(renderedLength(`${v}${'a'.repeat(140)}`, 20)).toBe(160)
    expect(checkContent({ ...ok, sms: `${v}${'a'.repeat(140)}` }, 20).sms).toBeUndefined()
    expect(checkContent({ ...ok, sms: `${v}${'a'.repeat(141)}` }, 20).sms).toBe('Too long after the name is filled in.')
    expect(checkContent({ ...ok, email_subject: `${v}${'b'.repeat(60)}` }, 20).email_subject).toBeUndefined()
    expect(checkContent({ ...ok, email_subject: `${v}${'b'.repeat(61)}` }, 20).email_subject).toBe('Too long after the name is filled in.')
    expect(checkContent({ ...ok, sms: `${v}${'ş'.repeat(140)}` }, 20).sms).toBeUndefined()
  })

  it('gives 160 for 146 runes plus one variable at 14 (FR-30)', () => {
    expect(renderedLength(`{{first_name}}${'a'.repeat(146)}`, 14)).toBe(160)
    expect(renderedLength(`{{first_name}}${'a'.repeat(147)}`, 14)).toBe(161)
  })

  it('accepts valid content', () => {
    expect(checkContent(ok, 14)).toEqual({})
  })

  it.each([
    ['empty', { ...ok, sms: '   ' }, 'sms', 'Write some text.'],
    ['wrong variable', { ...ok, sms: 'Hi {{last_name}}' }, 'sms', 'Only {{first_name}} is allowed.'],
    ['single brace', { ...ok, email_body: 'Hi {first_name}' }, 'email_body', 'Only {{first_name}} is allowed.'],
    ['spaced variable', { ...ok, sms: 'Hi {{ first_name }}' }, 'sms', 'Only {{first_name}} is allowed.'],
    ['subject newline', { ...ok, email_subject: 'a\nb' }, 'email_subject', 'The subject must be one line.'],
    ['html', { ...ok, email_body: '<b>x</b>' }, 'email_body', 'HTML is not allowed.'],
    ['control char', { ...ok, sms: 'a\u0007b' }, 'sms', 'Remove the special characters.'],
    ['newline in sms', { ...ok, sms: 'a\nb' }, 'sms', 'Remove the special characters.'],
  ] as const)('rejects %s', (_name, c, field, message) => {
    expect(checkContent(c, 14)[field]).toBe(message)
  })

  it('allows newline and tab in the body and a lone < in text', () => {
    expect(checkContent({ ...ok, email_body: 'a\n\tb < 3' }, 14)).toEqual({})
  })
})
