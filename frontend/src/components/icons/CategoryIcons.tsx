/**
 * Realistic two-tone blue SVG icons for the «Өз бағытыңызды табыңыз» section.
 * Style mirrors the "Browse by category" screenshot: detailed line-art filled
 * with light-blue fill areas and darker blue strokes/details.
 */

const PRIMARY = "#2563eb";   // stroke / dark detail
const LIGHT   = "#bfdbfe";   // fill / light area

interface IconProps {
  size?: number;
  "aria-hidden"?: "true";
  className?: string;
}

/** Сату (Sales) — bar chart with upward arrow */
export function SalesIcon({ size = 32, ...rest }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg" strokeLinecap="round" strokeLinejoin="round" {...rest}>
      {/* bars */}
      <rect x="6"  y="28" width="8" height="14" rx="3" fill={LIGHT} stroke={PRIMARY} strokeWidth="2"/>
      <rect x="20" y="18" width="8" height="24" rx="3" fill={LIGHT} stroke={PRIMARY} strokeWidth="2"/>
      <rect x="34" y="10" width="8" height="32" rx="3" fill={LIGHT} stroke={PRIMARY} strokeWidth="2"/>
      {/* trend arrow */}
      <path d="M7 32 20 22l14-8 6-4M34 10h6v6" stroke={PRIMARY} strokeWidth="2.2"/>
    </svg>
  );
}

/** Инженерия (Engineering) — gear / cog */
export function EngineeringIcon({ size = 32, ...rest }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg" strokeLinecap="round" {...rest}>
      <circle cx="24" cy="24" r="13" fill={LIGHT} stroke={PRIMARY} strokeWidth="2" />
      <path d="M24 6v5M24 37v5M6 24h5M37 24h5M11.3 11.3l3.5 3.5M33.2 33.2l3.5 3.5M36.7 11.3l-3.5 3.5M14.8 33.2l-3.5 3.5" stroke={PRIMARY} strokeWidth="3" />
      <circle cx="24" cy="24" r="5" fill="white" stroke={PRIMARY} strokeWidth="2" />
    </svg>
  );
}

/** Қаржы (Finance) — coin stack with dollar */
export function FinanceIcon({ size = 32, ...rest }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg" strokeLinecap="round" strokeLinejoin="round" {...rest}>
      {/* bottom coin */}
      <ellipse cx="24" cy="38" rx="14" ry="5" fill={LIGHT} stroke={PRIMARY} strokeWidth="2"/>
      <path d="M10 38v3c0 2.76 6.27 5 14 5s14-2.24 14-5v-3" stroke={PRIMARY} strokeWidth="2"/>
      {/* middle coin */}
      <ellipse cx="24" cy="28" rx="14" ry="5" fill={LIGHT} stroke={PRIMARY} strokeWidth="2"/>
      <path d="M10 28v3c0 2.76 6.27 5 14 5s14-2.24 14-5v-3" stroke={PRIMARY} strokeWidth="2"/>
      {/* top coin */}
      <ellipse cx="24" cy="18" rx="14" ry="5" fill="white" stroke={PRIMARY} strokeWidth="2"/>
      <path d="M10 18v3c0 2.76 6.27 5 14 5s14-2.24 14-5v-3" stroke={PRIMARY} strokeWidth="2"/>
      {/* $ symbol */}
      <text x="24" y="22" textAnchor="middle" fontSize="9" fontWeight="700" fill={PRIMARY} fontFamily="sans-serif">$</text>
    </svg>
  );
}

/** IT / Әзірлеу — laptop with code brackets */
export function ITIcon({ size = 32, ...rest }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg" {...rest}>
      {/* laptop base */}
      <rect x="6" y="10" width="36" height="24" rx="3" fill={LIGHT} stroke={PRIMARY} strokeWidth="2"/>
      {/* screen inner */}
      <rect x="9" y="13" width="30" height="18" rx="1.5" fill="white" stroke={PRIMARY} strokeWidth="1.5"/>
      {/* code brackets */}
      <polyline points="16,18 12,22 16,26" stroke={PRIMARY} strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" fill="none"/>
      <polyline points="32,18 36,22 32,26" stroke={PRIMARY} strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" fill="none"/>
      <line x1="21" y1="16" x2="27" y2="28" stroke={PRIMARY} strokeWidth="2" strokeLinecap="round"/>
      {/* base bar */}
      <path d="M4 34h40l-2 4H6l-2-4z" fill={LIGHT} stroke={PRIMARY} strokeWidth="2" strokeLinejoin="round"/>
    </svg>
  );
}

/** Авиация (Aviation) — airplane */
export function AviationIcon({ size = 32, ...rest }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg" strokeLinecap="round" strokeLinejoin="round" {...rest}>
      <path d="M24 7c-1.8 0-3 1.5-3 3.5v12L8 29v2l13-1.5V36l-4 3v2l7-2 7 2v-2l-4-3v-6.5L40 31v-2l-13-6.5v-12C27 8.5 25.8 7 24 7z" fill={LIGHT} stroke={PRIMARY} strokeWidth="2" />
      <path d="M21 22.5h6" stroke={PRIMARY} strokeWidth="2" />
    </svg>
  );
}

/** Клиенттерді қолдау (Customer service) — headset */
export function CustomerServiceIcon({ size = 32, ...rest }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg" {...rest}>
      <path d="M9 26v-3C9 14.72 15.72 8 24 8s15 6.72 15 15v3" stroke={PRIMARY} strokeWidth="2.4" strokeLinecap="round" />
      <path d="M9 24h4a3 3 0 013 3v8a3 3 0 01-3 3h-2a4 4 0 01-4-4v-6a4 4 0 012-4zM39 24h-4a3 3 0 00-3 3v8a3 3 0 003 3h2a4 4 0 004-4v-6a4 4 0 00-2-4z" fill={LIGHT} stroke={PRIMARY} strokeWidth="2" />
      <path d="M32 38c-1.5 2-4.25 3-8.25 3" stroke={PRIMARY} strokeWidth="2" strokeLinecap="round" />
      <rect x="19" y="38" width="7" height="5" rx="2.5" fill={LIGHT} stroke={PRIMARY} strokeWidth="2" />
    </svg>
  );
}

/** Адам ресурстары (Human resources) — team */
export function HumanResourcesIcon({ size = 32, ...rest }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg" {...rest}>
      <circle cx="24" cy="16" r="6" fill={LIGHT} stroke={PRIMARY} strokeWidth="2" />
      <circle cx="11" cy="21" r="4" fill="white" stroke={PRIMARY} strokeWidth="2" />
      <circle cx="37" cy="21" r="4" fill="white" stroke={PRIMARY} strokeWidth="2" />
      <path d="M13 40v-3c0-6.08 4.92-11 11-11s11 4.92 11 11v3H13z" fill={LIGHT} stroke={PRIMARY} strokeWidth="2" strokeLinejoin="round" />
      <path d="M5 38v-2a7 7 0 017-7M43 38v-2a7 7 0 00-7-7" stroke={PRIMARY} strokeWidth="2" strokeLinecap="round" />
    </svg>
  );
}

/** Маркетинг (Marketing) — megaphone */
export function MarketingIcon({ size = 32, ...rest }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 48 48" fill="none" xmlns="http://www.w3.org/2000/svg" {...rest}>
      <path d="M8 21v10a3 3 0 003 3h7l18 7V11l-18 7h-7a3 3 0 00-3 3z" fill={LIGHT} stroke={PRIMARY} strokeWidth="2" strokeLinejoin="round" />
      <path d="M18 34l3 9h7l-4-7.5M40 20l4-3M40 32l4 3M41 26h5" stroke={PRIMARY} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
      <path d="M13 22v8" stroke="white" strokeWidth="2" strokeLinecap="round" />
    </svg>
  );
}
