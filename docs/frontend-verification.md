# Frontend verification

Frontend changes should use the smallest check that proves the touched behavior,
then the standard package checks before handoff.

## Routine checks

```sh
cd frontend
npm run lint
npm run typecheck
npm run build
```

Run these after changes to React components, repositories, DTO mappers,
i18n keys, route wiring, styles or build configuration.

## Browser checks

For UI changes, verify the affected route in a browser at a desktop width and a
narrow mobile width. Look for:

- text wrapping inside buttons, form controls, filters and cards;
- empty, loading and error states;
- keyboard focus movement through interactive controls;
- Kazakh and Russian strings fitting in the same layout;
- external vacancies keeping external apply behavior.

Do not use English browser QA for the current MVP unless English support is
explicitly restored later.
