import type { JobSummary } from "../api/models/job";

export const publicationStatuses = [
  "published",
  "draft",
  "paused",
  "closed",
] as const;

export function jobUpdatedTime(job: JobSummary) {
  const date = Date.parse(job.updatedAt || job.createdAt || job.postedAt);
  return Number.isNaN(date) ? 0 : date;
}
