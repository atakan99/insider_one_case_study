// @vitest-environment happy-dom

import { describe, expect, it } from 'vitest'

import { buildPoints, SEND_LIMIT } from '../throughput-chart.vue'

describe('throughput chart data', () => {
  it('builds one point per bucket with the limit', () => {
    const sms = Array.from({ length: 120 }, (_, i) => i)
    const email = Array.from<number>({ length: 120 }).fill(2)
    const pts = buildPoints({ bucket_seconds: 5, start: '2026-11-10T12:00:00Z', sms, email })
    expect(pts).toHaveLength(120)
    expect(pts[0]).toEqual({ x: Date.parse('2026-11-10T12:00:00Z'), sms: 0, email: 2, limit: 100 })
    expect(pts[119].x).toBe(Date.parse('2026-11-10T12:00:00Z') + 119 * 5000)
    expect(pts[119].sms).toBe(119)
    expect(SEND_LIMIT).toBe(100)
  })
})
