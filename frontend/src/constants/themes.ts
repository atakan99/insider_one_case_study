export const THEMES = ['zinc', 'red', 'rose', 'orange', 'green', 'blue', 'yellow', 'violet'] as const
export type Theme = typeof THEMES[number]

export type Radius = typeof RADIUS[number]
export const RADIUS = [0, 0.25, 0.5, 0.75, 1] as const

export type ContentLayout = 'full' | 'centered'
