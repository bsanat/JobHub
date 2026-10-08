import { Check, ChevronDown } from "lucide-react";
import { useEffect, useId, useRef, useState, type KeyboardEvent } from "react";
import { useTranslation } from "react-i18next";

const languages = [
  { id: "kk", name: "Қазақша", shortName: "Қаз", flag: "kz" },
  { id: "ru", name: "Русский", shortName: "Рус", flag: "ru" },
] as const;

export function LanguageSelect() {
  const { t, i18n } = useTranslation();
  const language = i18n.resolvedLanguage === "ru" ? "ru" : "kk";
  const selected = languages.find((item) => item.id === language)!;
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const menuRef = useRef<HTMLDivElement>(null);
  const menuId = useId();

  useEffect(() => {
    if (!open) return;
    menuRef.current?.querySelector<HTMLButtonElement>('[aria-checked="true"]')?.focus();
    const dismiss = (event: PointerEvent) => {
      if (!rootRef.current?.contains(event.target as Node)) setOpen(false);
    };
    document.addEventListener("pointerdown", dismiss);
    return () => document.removeEventListener("pointerdown", dismiss);
  }, [open]);

  const onMenuKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key === "Escape") {
      event.preventDefault();
      setOpen(false);
      triggerRef.current?.focus();
      return;
    }
    const buttons = Array.from(menuRef.current?.querySelectorAll<HTMLButtonElement>("button") ?? []);
    const index = buttons.indexOf(document.activeElement as HTMLButtonElement);
    const next = event.key === "Home" ? 0 : event.key === "End" ? buttons.length - 1
      : event.key === "ArrowDown" ? (index + 1) % buttons.length
      : event.key === "ArrowUp" ? (index + buttons.length - 1) % buttons.length : -1;
    if (next >= 0) {
      event.preventDefault();
      buttons[next]?.focus();
    }
  };

  return (
    <div className="language-select" ref={rootRef} onBlur={(event) => {
      if (!event.currentTarget.contains(event.relatedTarget)) setOpen(false);
    }}>
      <button ref={triggerRef} className="language-select__trigger" type="button"
        aria-label={`${t("common.language")}: ${selected.name}`} aria-haspopup="menu"
        aria-expanded={open} aria-controls={open ? menuId : undefined}
        onClick={() => setOpen((value) => !value)}
        onKeyDown={(event) => {
          if (event.key === "ArrowDown" || event.key === "ArrowUp") {
            event.preventDefault();
            setOpen(true);
          }
        }}>
        <img className="language-select__flag" src={`/flags/${selected.flag}.svg`} alt="" width="20" height="15" />
        <span>{selected.shortName}</span>
        <ChevronDown size={14} aria-hidden="true" />
      </button>
      {open ? (
        <div ref={menuRef} id={menuId} className="language-select__menu" role="menu" aria-label={t("common.language")} onKeyDown={onMenuKeyDown}>
          {languages.map((item) => (
            <button key={item.id} type="button" role="menuitemradio" aria-checked={item.id === language}
              onClick={() => {
                void i18n.changeLanguage(item.id);
                setOpen(false);
                triggerRef.current?.focus();
              }}>
              <img className="language-select__flag" src={`/flags/${item.flag}.svg`} alt="" width="20" height="15" />
              <span>{item.name}</span>
              {item.id === language ? <Check size={16} aria-hidden="true" /> : null}
            </button>
          ))}
        </div>
      ) : null}
    </div>
  );
}
