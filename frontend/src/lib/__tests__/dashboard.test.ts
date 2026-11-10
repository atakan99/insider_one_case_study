import { describe, expect, it } from 'vitest'

import { formatDuration, percent, successRate } from '../dashboard'

const z = { sent: 0, failed: 0, unknown: 0, skipped: 0, in_progress: 0 }

describe('dashboard helpers', () => {
  it('computes the success rate over sent, failed and unknown', () => {
    expect(successRate({ ...z, sent: 95, unknown: 5 })).toBe('95.0%')
  })

  it('shows a dash when nothing finished', () => {
    expect(successRate(z)).toBe('–')
  })

  it('formats a percentage', () => {
    expect(percent(1, 3)).toBe('33.3%')
    expect(percent(1, 0)).toBe('–')
  })

  it('formats durations', () => {
    expect(formatDuration(3120)).toBe('3.1 s')
    expect(formatDuration(125000)).toBe('2 min 5 s')
    expect(formatDuration(null)).toBe('–')
  })
})
