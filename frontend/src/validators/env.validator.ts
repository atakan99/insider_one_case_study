import z from 'zod'

// Defaults match .env.example, so `pnpm dev` works on a clean clone without .env.
// A value that is set but invalid (for example a bad URL) still fails.
export const EnvSchema = z.object({
  VITE_SERVER_API_URL: z.url().default('http://localhost:8080'),
  VITE_SERVER_API_PREFIX: z.string().default('/api'),
  VITE_SERVER_API_TIMEOUT: z.coerce.number().default(10000),
})

export type env = z.infer<typeof EnvSchema>
