# Seed demo notes

`make seed-demo` loads local demo data through the Go seed command.

## Intended use

- Local product walkthroughs.
- Browser QA for candidate and employer happy paths.
- Empty-versus-populated UI comparisons.
- Development environments where synthetic data is acceptable.

## Safety rules

- Do not run seed commands against staging or production.
- Do not mix demo data with private resumes or real account data.
- Keep demo credentials and names synthetic.
- Re-run local migrations first when schema changes affect seeded tables.

When a feature depends on demo data, document the expected route and account role
in the PR or handoff.
