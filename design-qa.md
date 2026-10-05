# Jobs Page Visual QA

Date: 2026-10-05

## Latest Iteration: Salary Slider And Offer Rows

- Current references: `/var/folders/z_/14_ks9592jg5j45f2d5lqlqh0000gn/T/TemporaryItems/NSIRD_screencaptureui_NS3k57/Снимок экрана 2026-10-05 в 21.48.14.png` (salary control) and `/var/folders/z_/14_ks9592jg5j45f2d5lqlqh0000gn/T/TemporaryItems/NSIRD_screencaptureui_ITJJTz/Снимок экрана 2026-10-05 в 21.49.29.png` (offer rows).
- Current evidence: `/tmp/jobhub-offers-desktop.jpg` (1440 x 1000 CSS/image pixels), `/tmp/jobhub-offers-mobile.jpg` (390 x 844), `/tmp/jobhub-salary-mobile.jpg` (390 x 844). No density scaling; compare individual reference rows and the salary region, since the supplied images are cropped components rather than full application viewports.
- Full-view comparison: source offer screenshot and final desktop screenshot inspected together. Title/skills occupy the left column; salary/company/location/save occupy the right. Source flame, viewed and new badges are omitted because equivalent job state is unavailable. Existing JobHub header, filter rail and blue palette remain intentional product constraints.
- Focused comparison: desktop and mobile salary control inspected against the supplied slider reference. Two draggable handles, visible bounds, currency selector and Apply action fit without overlap. Native keyboard interaction retained.
- Typography: existing fonts retained, 18px titles and salaries, 12px skills, 14px metadata. Long values wrap; no clamped vacancy titles.
- Spacing: compact two-row desktop layout, stacked mobile cards. No horizontal overflow at 1440px or 390px. Cards remain within the existing 8px radius convention.
- Colors/assets: existing blue tokens and company logos retained; Kaspi fills its circular avatar, wide wordmarks remain complete. Save uses the existing persistence/auth behavior with a heart icon.
- Copy: Kazakh and Russian salary labels added. Real source attribution retained. No invented urgency or salary amounts.
- Interactions: dragged upper handle to 880,000 KZT; applied 300,000-880,000 and observed three overlapping salary ranges; upper-only 590,000 returned one matching fixture. Refresh preserved the upper bound. Changing currency reset the draft, applying the full range cleared the salary URL parameters. Mobile Apply and reset also verified.
- Range semantics: monthly salary intervals overlap the selected interval, in the selected currency; unknown/private salaries are excluded when a bound is active. Frontend URL, HTTP, mock repository, handler, count/list SQL and generated code support both bounds.
- Checks: frontend lint/typecheck/build passed; Go tests, vet and build passed. sqlc v1.30.0 regenerated queries. Invalid, non-finite and reversed bounds have handler regression coverage. Database integration assertions were added but skipped because no isolated TEST_DATABASE_URL is configured. Preview uses fixture data, not a live database. Browser console: no errors.
- Final comparison found no actionable visual P0/P1/P2 issues within this scope. Prior comparison and fixes are retained below as history.

## Previous Iteration

- Source: `/var/folders/z_/14_ks9592jg5j45f2d5lqlqh0000gn/T/codex-clipboard-0ea647ab-c3b6-45f0-82fe-dc376da9870b.png` (2880 x 1800).
- Implementation: `/tmp/jobhub-vacancies-preview.jpg` (1280 x 900); mobile: `/tmp/jobhub-jobs-mobile.png` (390 x 844).
- Route: `http://localhost:5174/jobs`, existing mock repository, signed out.
- Viewports checked: 1440 x 1000, 1280 x 900, 390 x 844 CSS pixels. Saved browser captures use one image pixel per CSS pixel.
- Comparison: reference browser chrome and surrounding Dribbble canvas excluded conceptually; compared the central filter rail and vacancy list, rather than treating the reduced embedded mockup as a full-size application viewport.

## Findings And Fixes

- Initial P2: search field had a double border. Removed the wrapper border; the input retains its focus indication. Final capture shows one border.
- Initial P2: uniform cover sizing clipped the wide Kolesa wordmark. CompanyLogo now contains wide images while square marks fill the circular mask. Final capture shows the complete wordmark and edge-to-edge Kaspi mark.
- Fonts: existing JobHub fonts retained; compact card titles use 18px, metadata uses 14px, dates use 12px. No title clamping in compact cards.
- Layout: left search/filter rail, results count and sorting above horizontal cards; logo and identity aligned, metadata below, action at lower right. Mobile uses the existing filter drawer and full-width card actions. No horizontal overflow at checked widths.
- Colors: existing primary, surface, text, and border tokens retained as requested; rail and date tint derive from the primary token.
- Assets: existing company assets retained. Missing assets use the existing building icon. Focused logo inspection confirmed loaded images, zero avatar padding, cover for Kaspi, contain for the wide wordmark.
- Content: existing Kazakh/Russian UI and job data retained. Imported-job source remains visible and its detail route is preserved. Existing filter contracts retained; reference-only category counts, tags, salary range, and hiring advertisement were not invented.

## Verification

- Junior filter reduced five fixture jobs to one; reset restored the list.
- Mobile drawer opened; searching Kaspi closed the drawer and returned its one fixture vacancy.
- Both supported locales rendered; preview console reported no errors.
- Lint, TypeScript check, production build, and diff whitespace check passed.
- Live API integration remains unverified: the original port 5173 page showed a load error before the redesign. Mock preview does not alter API configuration.

final result: passed
