import { describe, expect, it } from 'vitest'

import { formatCount, formatDateTime } from '../format'

describe('format', () => {
  it('formats a count with thousands separators', () => {
    expect(formatCount(1212)).toBe('1,212')
    expect(formatCount(0)).toBe('0')
  })

  it('formats a date in English and returns an empty string for a bad date', () => {
    expect(formatDateTime('2026-11-10T12:00:00Z')).toContain('Nov 10, 2026')
    expect(formatDateTime('not a date')).toBe('')
  })
})
