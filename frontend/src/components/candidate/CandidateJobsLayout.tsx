import type { PropsWithChildren, ReactNode } from "react";
import { CandidateWorkspaceLayout } from "./CandidateWorkspaceLayout";

interface CandidateJobsLayoutProps extends PropsWithChildren {
  tabTitle: string;
  action?: ReactNode;
}

export function CandidateJobsLayout({
  children,
  tabTitle,
  action,
}: CandidateJobsLayoutProps) {
  return (
    <CandidateWorkspaceLayout>
      <section
        className="candidate-list-page"
        aria-labelledby="candidate-jobs-title"
      >
        <header className="workbench-page-header">
          <div>
            <h1 id="candidate-jobs-title">{tabTitle}</h1>
          </div>
          {action}
        </header>
        {children}
      </section>
    </CandidateWorkspaceLayout>
  );
}
