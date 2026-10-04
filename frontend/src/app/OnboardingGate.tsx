import type { PropsWithChildren } from "react";
import { Navigate, useLocation } from "react-router-dom";
import { useAuth } from "./authContext";
import { onboardingDestination, onboardingPath } from "../lib/onboarding";

export function OnboardingGate({ children }: PropsWithChildren) {
  const { session } = useAuth();
  const location = useLocation();
  if (session?.onboardingRequired && location.pathname !== "/onboarding") {
    const onAuthPage =
      location.pathname === "/login" || location.pathname === "/register";
    const returnTo =
      location.pathname === "/register" && session.user.role === "employer"
        ? "/employer"
        : onAuthPage
          ? new URLSearchParams(location.search).get("returnTo")
          : location.pathname + location.search + location.hash;
    const destination = onboardingDestination(returnTo, session.user.role);
    return <Navigate to={onboardingPath(destination)} replace />;
  }
  return children;
}
