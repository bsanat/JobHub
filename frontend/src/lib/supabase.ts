import { createClient } from "@supabase/supabase-js";
import { env } from "./env";

export const googleAuthConfigured = Boolean(env.supabaseUrl && env.supabaseAnonKey);

export const supabase = googleAuthConfigured
  ? createClient(env.supabaseUrl!, env.supabaseAnonKey!, {
      auth: {
        persistSession: true,
        autoRefreshToken: true,
        detectSessionInUrl: true,
      },
    })
  : null;

export function googleRedirectTo(returnTo: string | null) {
  const url = new URL("/auth/callback", window.location.origin);
  if (returnTo) url.searchParams.set("returnTo", returnTo);
  return url.toString();
}
