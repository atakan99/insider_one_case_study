import type { Content } from '@/lib/content'

// The same limits and checks as the backend validator (internal/content).
// The backend decides; this check only helps the marketer while typing.
export const LIMITS = { sms: 160, email_subject: 80, email_body: 2000 } as const

const VAR = '{{first_name}}'
const TAG = /<[a-z/!]/i

export function runeLength(s: string): number {
  return [...s].length
}

// The length after each {{first_name}} is replaced by a name of firstNameRunes runes.
export function renderedLength(text: string, firstNameRunes: number): number {
  const vars = text.split(VAR).length - 1
  return runeLength(text) - vars * runeLength(VAR) + vars * firstNameRunes
}

function hasControl(s: string, allowBody: boolean): boolean {
  for (const ch of s) {
    const c = ch.codePointAt(0)!
    if (allowBody && (ch === '\n' || ch === '\t'))
      continue
    if (c < 0x20 || (c >= 0x7F && c <= 0x9F))
      return true
  }
  return false
}

export function checkContent(c: Content, firstNameRunes: number): Partial<Record<keyof Content, string>> {
  const out: Partial<Record<keyof Content, string>> = {}
  for (const key of ['sms', 'email_subject', 'email_body'] as const) {
    const text = c[key].trim()
    if (text === '') {
      out[key] = 'Write some text.'
      continue
    }
    const rest = text.split(VAR).join('')
    if (rest.includes('{') || rest.includes('}'))
      out[key] = 'Only {{first_name}} is allowed.'
    else if (key === 'email_subject' && /[\r\n]/.test(text))
      out[key] = 'The subject must be one line.'
    else if (hasControl(text, key === 'email_body'))
      out[key] = 'Remove the special characters.'
    else if (TAG.test(text))
      out[key] = 'HTML is not allowed.'
    else if (renderedLength(text, firstNameRunes) > LIMITS[key])
      out[key] = 'Too long after the name is filled in.'
  }
  return out
}
