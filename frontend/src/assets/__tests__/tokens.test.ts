import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'
import { describe, expect, it } from 'vitest'

const css = readFileSync(resolve(__dirname, '../index.css'), 'utf8')

function block(start: string): string {
  const from = css.indexOf(start)
  expect(from).toBeGreaterThanOrEqual(0)
  return css.slice(from, css.indexOf('\n}', from))
}

describe('outcome colour tokens', () => {
  for (const token of ['success', 'success-foreground', 'info']) {
    it(`--${token} exists in both themes and in the Tailwind theme`, () => {
      expect(block(':root {')).toMatch(new RegExp(`--${token}: oklch\\(`))
      expect(block('.dark {')).toMatch(new RegExp(`--${token}: oklch\\(`))
      expect(block('@theme inline {')).toContain(`--color-${token}: var(--${token});`)
    })
  }
})
