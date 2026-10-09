import { createAuthClient } from "@neondatabase/auth";
import { env } from "./env";

export const neonAuthConfigured = Boolean(env.neonAuthUrl);
export const googleAuthConfigured = neonAuthConfigured;
export const neonAuth = env.neonAuthUrl ? createAuthClient(env.neonAuthUrl) : null;

export function googleRedirectTo(returnTo: string | null) {
 const url = new URL("/auth/callback", window.location.origin);
 if (returnTo) url.searchParams.set("returnTo", returnTo);
 return url.toString();
}
