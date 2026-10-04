import type { UserRole } from "../api/models/auth";
import { safeReturnTo } from "./returnTo";

export function accountDestination(role: UserRole) {
  return role === "employer"
    ? "/employer"
    : role === "admin"
      ? "/admin"
      : "/profile";
}

export function onboardingDestination(value: string | null, role: UserRole) {
  const fallback = accountDestination(role);
  const path = safeReturnTo(value, fallback);
  let url: URL;
  try {
    url = new URL(path, window.location.origin);
  } catch {
    return fallback;
  }
  if (
    url.origin !== window.location.origin ||
    /^\/(onboarding|login|register)(\/|$)/.test(url.pathname) ||
    url.pathname === "/auth/callback"
  )
    return fallback;
  return url.pathname + url.search + url.hash;
}

export function onboardingPath(destination: string) {
  return `/onboarding?returnTo=${encodeURIComponent(destination)}`;
}
