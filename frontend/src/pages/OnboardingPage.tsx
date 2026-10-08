import { useQuery, useQueryClient } from "@tanstack/react-query";
import { ArrowLeft, ArrowRight, Check, LogOut } from "lucide-react";
import { useEffect, useRef, useState, type FormEvent } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate, useSearchParams } from "react-router-dom";
import type {
  AuthSession,
  OnboardingAnswers,
  OnboardingState,
} from "../api/models/auth";
import { repositories } from "../api/repositories";
import { useAuth } from "../app/authContext";
import { LanguageSelect } from "../components/i18n/LanguageSelect";
import { CitySelect } from "../components/location/CitySelect";
import { Button } from "../components/ui/Button";
import { Input } from "../components/ui/FormControls";
import { JobHubLoader } from "../components/ui/JobHubLoader";
import { onboardingDestination } from "../lib/onboarding";
import { authErrorCode } from "./authUtils";
import "../styles/onboarding.css";

const personas = [
  "school_student",
  "university_student",
  "graduate",
  "unemployed",
  "other",
] as const;
const goals = ["job", "internship", "browsing"] as const;
const referrals = [
  "telegram",
  "instagram",
  "google",
  "friend",
  "university",
  "other",
] as const;

export function OnboardingPage() {
  const { session, logout } = useAuth();
  const { t, i18n } = useTranslation();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const query = useQuery({
    queryKey: ["auth", "onboarding", session?.user.id],
    queryFn: () => repositories.auth.onboarding(),
    enabled: Boolean(session),
    staleTime: Infinity,
    refetchOnWindowFocus: false,
    retry: false,
  });
  const destination = onboardingDestination(
    params.get("returnTo"),
    session?.user.role ?? "job_seeker",
  );

  const finish = () => {
    queryClient.setQueryData<AuthSession>(["auth", "me"], (current) =>
      current && current.user.id === session?.user.id
        ? { ...current, onboardingRequired: false }
        : current,
    );
    void queryClient.invalidateQueries({
      queryKey: ["auth", "onboarding", session?.user.id],
    });
    void navigate(destination, { replace: true });
  };

  useEffect(() => {
    if (query.data && !query.data.required) {
      queryClient.setQueryData<AuthSession>(["auth", "me"], (current) =>
        current && current.user.id === session?.user.id
          ? { ...current, onboardingRequired: false }
          : current,
      );
      void navigate(destination, { replace: true });
    }
  }, [query.data, queryClient, session?.user.id, destination, navigate]);

  useEffect(() => {
    document.documentElement.lang =
      i18n.resolvedLanguage === "ru" ? "ru" : "kk";
  }, [i18n.resolvedLanguage]);

  return (
    <div className="onboarding-shell">
      <a className="skip-link" href="#main-content">
        {t("common.skipToContent")}
      </a>
      <header className="onboarding-header">
        <img
          src="/brand/jobhub-logo.png"
          width="600"
          height="187"
          alt="JobHub"
        />
        <LanguageSelect />
      </header>
      <main id="main-content" className="onboarding-main">
        {!query.data || !query.data.required ? (
          query.isError ? (
            <section className="onboarding-unavailable">
              <h1>{t("onboarding.title")}</h1>
              <p role="alert">{t("onboarding.loadError")}</p>
              <Button onClick={() => void query.refetch()}>
                {t("common.retry")}
              </Button>
              <Button
                variant="quiet"
                leadingIcon={<LogOut size={18} />}
                onClick={() =>
                  void logout().then(() =>
                    navigate("/login", { replace: true }),
                  )
                }
              >
                {t("auth.logout")}
              </Button>
            </section>
          ) : (
            <JobHubLoader />
          )
        ) : session ? (
          <OnboardingForm
            key={`${session.user.id}:${query.data.version}`}
            initial={query.data}
            session={session}
            onFinish={finish}
            onReload={() => void query.refetch()}
          />
        ) : null}
      </main>
    </div>
  );
}

function OnboardingForm({
  initial,
  session,
  onFinish,
  onReload,
}: {
  initial: OnboardingState;
  session: AuthSession;
  onFinish: () => void;
  onReload: () => void;
}) {
  const { t } = useTranslation();
  const { logout } = useAuth();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [draft, setDraft] = useState<OnboardingAnswers>(initial);
  const [saveStatus, setSaveStatus] = useState("saved");
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const [validation, setValidation] = useState("");
  const heading = useRef<HTMLHeadingElement>(null);
  const draftRef = useRef<OnboardingAnswers>(initial);
  const version = useRef(initial.version);
  const saved = useRef(JSON.stringify(initial));
  const conflicted = useRef(false);
  const queue = useRef<Promise<boolean>>(Promise.resolve(true));
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);

  useEffect(() => {
    heading.current?.focus();
  }, [draft.step]);
  useEffect(() => {
    const warn = (event: BeforeUnloadEvent) => {
      if (JSON.stringify(draftRef.current) !== saved.current) {
        event.preventDefault();
        event.returnValue = "";
      }
    };
    window.addEventListener("beforeunload", warn);
    return () => {
      clearTimeout(timer.current);
      window.removeEventListener("beforeunload", warn);
    };
  }, []);

  // Serialize saves so typing, blur and navigation cannot race with each other.
  const persist = (snapshot: OnboardingAnswers, complete = false) => {
    const signature = JSON.stringify(snapshot);
    queue.current = queue.current.then(async () => {
      if (conflicted.current) return false;
      if (!complete && signature === saved.current) return true;
      setSaveStatus("saving");
      try {
        const state = await repositories.auth.saveOnboarding(
          snapshot,
          version.current,
          complete,
          session.csrfToken,
        );
        version.current = state.version;
        saved.current = signature;
        setError("");
        setSaveStatus(
          JSON.stringify(draftRef.current) === signature ? "saved" : "unsaved",
        );
        if (!state.required) onFinish();
        return true;
      } catch (cause) {
        const code = authErrorCode(cause);
        conflicted.current = code === "ONBOARDING_CONFLICT";
        setError(
          conflicted.current
            ? "conflict"
            : code === "UNAUTHENTICATED" || code === "CSRF_INVALID"
              ? "sessionError"
              : "saveError",
        );
        setSaveStatus("unsaved");
        return false;
      }
    });
    return queue.current;
  };

  const update = (patch: Partial<OnboardingAnswers>) => {
    const next = { ...draftRef.current, ...patch };
    draftRef.current = next;
    setDraft(next);
    setValidation("");
    setSaveStatus("unsaved");
    clearTimeout(timer.current);
    timer.current = setTimeout(() => void persist(next), 350);
  };
  const flush = () => {
    clearTimeout(timer.current);
    return persist(draftRef.current);
  };

  const advance = async (event?: FormEvent, skip = false) => {
    event?.preventDefault();
    const current = draftRef.current;
    const missing =
      current.step === 1
        ? !current.persona ||
          (current.persona === "other" && !current.personaOther.trim())
        : current.step === 2
          ? !current.goal
          : current.step === 3
            ? !current.cityId
            : current.step === 4
              ? !current.profession.trim()
              : !skip &&
                current.referral === "other" &&
                !current.referralOther.trim();
    if (missing) {
      setValidation(
        current.step === 1 || current.step === 5
          ? "answerRequired"
          : "required",
      );
      return;
    }
    clearTimeout(timer.current);
    const next = {
      ...current,
      step: Math.min(5, current.step + 1),
      ...(skip ? { referral: "skip" as const, referralOther: "" } : {}),
    };
    setBusy(true);
    const ok = await persist(next, current.step === 5);
    if (ok && current.step < 5) {
      draftRef.current = next;
      setDraft(next);
      setValidation("");
      setSaveStatus("saved");
    }
    setBusy(false);
  };

  const back = async () => {
    clearTimeout(timer.current);
    const next = { ...draftRef.current, step: draftRef.current.step - 1 };
    setBusy(true);
    if (await persist(next)) {
      draftRef.current = next;
      setDraft(next);
      setValidation("");
      setSaveStatus("saved");
    }
    setBusy(false);
  };

  const leave = async () => {
    setBusy(true);
    if (await flush()) {
      try {
        await logout();
        void navigate("/login", { replace: true });
      } catch {
        setError("sessionError");
      }
    }
    setBusy(false);
  };

  const inlineError = validation ? t(`onboarding.${validation}`) : undefined;
  const otherQuestion =
    (draft.step === 1 && draft.persona === "other") ||
    (draft.step === 5 && draft.referral === "other");

  return (
    <section className="onboarding-flow" aria-labelledby="onboarding-title">
      <header className="onboarding-intro">
        <h1 id="onboarding-title">{t("onboarding.title")}</h1>
        <p>{t("onboarding.subtitle")}</p>
      </header>
      <div
        className="onboarding-progress"
        role="progressbar"
        aria-label={t("onboarding.progress")}
        aria-valuemin={1}
        aria-valuemax={5}
        aria-valuenow={draft.step}
        aria-valuetext={t("onboarding.step", { step: draft.step })}
      >
        <span>{t("onboarding.step", { step: draft.step })}</span>
        <div aria-hidden="true">
          {[1, 2, 3, 4, 5].map((step) => (
            <i key={step} className={step <= draft.step ? "is-complete" : ""} />
          ))}
        </div>
      </div>
      <form onSubmit={(event) => void advance(event)} noValidate>
        <h2 ref={heading} tabIndex={-1}>
          {t(`onboarding.questions.${draft.step}`)}
        </h2>
        {draft.step === 5 ? (
          <p className="onboarding-optional">{t("onboarding.optional")}</p>
        ) : null}
        <fieldset
          className="onboarding-fields"
          disabled={busy || error === "conflict"}
        >
          <legend className="visually-hidden">
            {t(`onboarding.questions.${draft.step}`)}
          </legend>
          {draft.step === 1 ? (
            <Choices
              name="persona"
              values={personas}
              value={draft.persona}
              namespace="personas"
              error={!otherQuestion ? inlineError : undefined}
              onChange={(persona) => update({ persona, personaOther: "" })}
            />
          ) : null}
          {draft.step === 2 ? (
            <Choices
              name="goal"
              values={goals}
              value={draft.goal}
              namespace="goals"
              error={inlineError}
              onChange={(goal) => update({ goal })}
            />
          ) : null}
          {draft.step === 3 ? (
            <CitySelect
              label={t("onboarding.city")}
              anyLabel={t("onboarding.cityPlaceholder")}
              value={draft.cityId}
              onChange={(cityId) => update({ cityId })}
              required
              error={inlineError}
            />
          ) : null}
          {draft.step === 4 ? (
            <Input
              label={t("onboarding.profession")}
              placeholder={t("onboarding.professionPlaceholder")}
              value={draft.profession}
              maxLength={200}
              required
              error={inlineError}
              onChange={(event) => update({ profession: event.target.value })}
              onBlur={() => void flush()}
            />
          ) : null}
          {draft.step === 5 ? (
            <Choices
              name="referral"
              values={referrals}
              value={draft.referral}
              namespace="referrals"
              onChange={(referral) => update({ referral, referralOther: "" })}
            />
          ) : null}
          {otherQuestion ? (
            <Input
              label={t("onboarding.otherAnswer")}
              value={
                draft.step === 1 ? draft.personaOther : draft.referralOther
              }
              maxLength={200}
              required
              error={inlineError}
              onChange={(event) =>
                update(
                  draft.step === 1
                    ? { personaOther: event.target.value }
                    : { referralOther: event.target.value },
                )
              }
              onBlur={() => void flush()}
            />
          ) : null}
        </fieldset>
        <div className="onboarding-save-status" role="status">
          {saveStatus === "saved" ? (
            <Check size={16} aria-hidden="true" />
          ) : null}
          {t(`onboarding.${saveStatus}`)}
        </div>
        {error ? (
          <div className="onboarding-error" role="alert">
            <p>{t(`onboarding.${error}`)}</p>
            <Button
              type="button"
              variant="secondary"
              size="sm"
              disabled={busy}
              onClick={() => {
                if (error === "conflict") onReload();
                else if (error === "sessionError") {
                  queryClient.setQueryData(["auth", "me"], null);
                  queryClient.removeQueries({
                    queryKey: ["auth", "onboarding"],
                  });
                  void navigate("/login", { replace: true });
                } else void flush();
              }}
            >
              {t(
                error === "conflict"
                  ? "onboarding.reload"
                  : error === "sessionError"
                    ? "auth.loginLink"
                    : "common.retry",
              )}
            </Button>
          </div>
        ) : null}
        <div className="onboarding-controls">
          <Button
            type="button"
            variant="secondary"
            leadingIcon={<ArrowLeft size={18} aria-hidden="true" />}
            disabled={draft.step === 1 || busy || error === "conflict"}
            onClick={() => void back()}
          >
            {t("onboarding.back")}
          </Button>
          <div>
            {draft.step === 5 ? (
              <Button
                type="button"
                variant="quiet"
                disabled={busy || error === "conflict"}
                onClick={() => void advance(undefined, true)}
              >
                {t("onboarding.skip")}
              </Button>
            ) : null}
            <Button
              type="submit"
              isLoading={busy}
              disabled={error === "conflict"}
            >
              {t(
                draft.step === 5 ? "onboarding.finish" : "onboarding.continue",
              )}
              <ArrowRight size={18} aria-hidden="true" />
            </Button>
          </div>
        </div>
      </form>
      <div className="onboarding-footer">
        <span>{draft.step} / 5</span>
        <Button
          variant="quiet"
          size="sm"
          disabled={busy}
          leadingIcon={<LogOut size={15} aria-hidden="true" />}
          onClick={() => void leave()}
        >
          {t("onboarding.leave")}
        </Button>
      </div>
    </section>
  );
}

function Choices<T extends string>({
  name,
  values,
  value,
  namespace,
  error,
  onChange,
}: {
  name: string;
  values: readonly T[];
  value: string;
  namespace: string;
  error?: string;
  onChange: (value: T) => void;
}) {
  const { t } = useTranslation();
  return (
    <div
      className="onboarding-choices"
      role="radiogroup"
      aria-label={t(
        `onboarding.questions.${name === "persona" ? 1 : name === "goal" ? 2 : 5}`,
      )}
      aria-invalid={Boolean(error)}
      aria-describedby={error ? `${name}-error` : undefined}
    >
      {values.map((option) => (
        <label key={option} className={value === option ? "is-selected" : ""}>
          <input
            type="radio"
            name={name}
            value={option}
            checked={value === option}
            onChange={() => onChange(option)}
          />
          <span>{t(`onboarding.${namespace}.${option}`)}</span>
        </label>
      ))}
      {error ? (
        <span id={`${name}-error`} className="ui-field-error">
          {error}
        </span>
      ) : null}
    </div>
  );
}
