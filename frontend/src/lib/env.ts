import { z } from "zod";

const browserEnvSchema = z.object({
  VITE_API_URL: z.union([z.string().url(), z.literal("/api/v1")]).default(
    import.meta.env.PROD ? "/api/v1" : "http://localhost:8080/api/v1",
  ),
  VITE_USE_MOCKS: z.enum(["true", "false"]).default(import.meta.env.PROD ? "false" : "true"),
  VITE_SUPABASE_URL: z.string().url().optional(),
  VITE_SUPABASE_ANON_KEY: z.string().min(1).optional(),
});

const parsedEnv = browserEnvSchema.parse({
  VITE_API_URL: import.meta.env.VITE_API_URL,
  VITE_USE_MOCKS: import.meta.env.VITE_USE_MOCKS,
  VITE_SUPABASE_URL: import.meta.env.VITE_SUPABASE_URL,
  VITE_SUPABASE_ANON_KEY: import.meta.env.VITE_SUPABASE_ANON_KEY,
});

export const env = {
  apiUrl: parsedEnv.VITE_API_URL.replace(/\/$/, ""),
  useMocks: parsedEnv.VITE_USE_MOCKS === "true",
  supabaseUrl: parsedEnv.VITE_SUPABASE_URL,
  supabaseAnonKey: parsedEnv.VITE_SUPABASE_ANON_KEY,
} as const;
