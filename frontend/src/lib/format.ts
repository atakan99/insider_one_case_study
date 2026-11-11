const DATE_TIME = new Intl.DateTimeFormat('en-US', { dateStyle: 'medium', timeStyle: 'short' })

// The UI is English only, so every count and date uses the en-US format.
export function formatCount(n: number): string {
  return n.toLocaleString('en-US')
}

export function formatDateTime(iso: string): string {
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? '' : DATE_TIME.format(d)
}
