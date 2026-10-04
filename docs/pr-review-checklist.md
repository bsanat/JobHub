# Pull request review checklist

Use this checklist for small JobHub pull requests.

## Scope

- The title describes the observable change.
- The PR avoids unrelated formatting, dependency upgrades and refactors.
- Documentation changes do not copy secrets, private data or stale reports.
- Generated files are updated only when their sources changed.

## Correctness

- Frontend changes use the typed API/repository layer.
- Backend authorization stays server-enforced.
- Native and imported vacancy behavior remains separate.
- New UI copy updates both Kazakh and Russian locale files.

## Verification

- Mention the focused checks that were run.
- For docs-only changes, state that runtime checks were not needed.
- Call out skipped checks honestly when tools or credentials are unavailable.
