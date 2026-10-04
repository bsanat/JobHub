import {
  Building2,
  BriefcaseBusiness,
  LayoutDashboard,
  Settings2,
  Users,
} from "lucide-react";

export const employerDestinations = [
  ["/employer", "overview", LayoutDashboard],
  ["/employer/vacancies", "vacancies", BriefcaseBusiness],
  ["/employer/applications", "applications", Users],
  ["/employer/company", "company", Building2],
  ["/account", "account", Settings2],
] as const;
