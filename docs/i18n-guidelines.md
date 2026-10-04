# i18n guidelines

JobHub currently ships Kazakh and Russian UI locales for the active MVP.
English is intentionally deferred.

## Current locale rule

- Add or update `kk` and `ru` keys for new user-facing UI.
- Do not add English files, English selector entries or English browser QA for
  the current MVP.
- Keep key names extensible so English can be restored later without reshaping
  the feature.

## Implementation notes

- Prefer feature-scoped keys over one-off text embedded in components.
- Keep labels, helper text, empty states and validation messages together when
  they belong to the same flow.
- Check longer Russian strings and Kazakh strings in narrow layouts.
- Do not reuse a key when two screens need different meaning or tone.

Backend error contracts should remain stable; frontend copy can translate the
user-facing message separately.
