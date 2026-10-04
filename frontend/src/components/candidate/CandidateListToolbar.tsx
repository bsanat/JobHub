import { Search, X } from "lucide-react";
import { useRef, type ReactNode } from "react";
import { useTranslation } from "react-i18next";
import { IconButton } from "../ui/Button";

interface CandidateListToolbarProps {
  query: string;
  onQueryChange: (value: string) => void;
  searchLabel: string;
  count: number;
  children?: ReactNode;
}

export function CandidateListToolbar({
  query,
  onQueryChange,
  searchLabel,
  count,
  children,
}: CandidateListToolbarProps) {
  const { t } = useTranslation();
  const input = useRef<HTMLInputElement>(null);
  return (
    <>
      <div className="candidate-list-toolbar">
        <div className="vacancy-search">
          <Search size={18} aria-hidden="true" />
          <input
            ref={input}
            type="search"
            aria-label={searchLabel}
            placeholder={searchLabel}
            value={query}
            onChange={(event) => onQueryChange(event.target.value)}
          />
          {query ? (
            <IconButton
              variant="quiet"
              label={t("candidateLists.clearSearch")}
              icon={<X size={16} />}
              onClick={() => {
                onQueryChange("");
                input.current?.focus();
              }}
            />
          ) : null}
        </div>
        {children}
      </div>
      <p className="vacancy-result-count" role="status">
        {t("candidateLists.resultCount", { count })}
      </p>
    </>
  );
}
