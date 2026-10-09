# JobHub: переход на Neon

Обновлено 9 октября 2026 года. Рабочий JobHub подключён к Neon PostgreSQL.
Это новая база с повторно импортированными публичными вакансиями,
а не восстановление полной старой Production-базы Supabase.

## Рабочая база и сайт

- Проект: `floral-frost-85848358`, ветка `production`
  (`br-sweet-bread-b5f9m2v5`).
- На проверенной пустой базе применены существующие миграции 1–15.
  Версия 15, `dirty=false`; все таблицы `jobhub` имеют RLS,
  все ограничения валидированы.
- После dry-run выполнен разовый импорт четырёх ранее разрешённых
  официальных источников:

| Источник | Вакансии | Официальная страница |
| --- | ---: | --- |
| Air Astana | 37 | https://job.airastana.com/ |
| Kcell | 39 | https://jobs.kcell.kz/ |
| Technodom | 38 | https://hr.technodom.kz/vacancies/ |
| Kolesa Group | 2 | https://kolesa.group/career/job |

- Всего 116 записей; одна запись Technodom пропущена нормализатором.
  Дубликатов `(source, external_id)` нет. Telegram не вызывался.
- Vercel Production использует `DATABASE_URL` (Neon pooled, Secret),
  `PGX_QUERY_EXEC_MODE=exec`, Neon Auth и приватное Neon Object Storage.
  Старые `SUPABASE_URL`, `SUPABASE_SECRET_KEY`, `VITE_SUPABASE_URL` и
  `VITE_SUPABASE_ANON_KEY` удалены из настроек проекта.
- Локальный `.env` переведён на изолированную ветку `jobhub-recovery-test`;
  старые настройки сохранены в приватном ignored `.env.before-neon.backup`.
  Импорты и Telegram в локальном окружении выключены.
  GitHub `PUBLIC_IMPORT_DATABASE_URL` и резервный `TELEGRAM_PUBLISHER_DATABASE_URL` переведены на Neon pooled; защитные
  `PUBLIC_IMPORT_DATABASE_HOST/USER` соответствуют проверенным hostname и роли.
  `PUBLIC_IMPORT_ENABLED`, `KZ_IMPORT_AUTO_ENABLED`, `TELEGRAM_PUBLISHER_ENABLED`
  остаются `false`; workflow не запускался.
- Deployment: `dpl_DrXufeERqagbVWShdDMTJE2Cnqc5`;
  сайт: https://jobhub-lovat.vercel.app/jobs.
- Старый deployment для отката публичного режима:
  `dpl_4Qii4snNs5n98YNUKeqcFvEMFCMh`.
- Перед переключением создана копия Neon `jobhub-before-cutover`
  (`br-weathered-breeze-b5524yzq`) без compute. Она сохраняет схему и данные
  перед переключением; это не внешний дамп и не копия Supabase.
  Локальный `pg_dump` не выполнен: установлен клиент PostgreSQL 16,
  сервер Neon — PostgreSQL 18. Пустой файл не сохранён.
- Seed-миграция автоматически ставит Air Astana `is_verified=true`.
  В новой рабочей базе этот флаг снят: у импортированной компании нет
  зарегистрированного владельца. У четырёх импортированных работодателей
  теперь нет verified badge. Применённые миграции не переписывались.

## Проверки рабочего подключения

До переключения проверен локальный Go API с Neon pooled connection и
новый deployment. После переключения проверены публичные health,
каталог, страницы пагинации, региональный фильтр, подробная карточка,
атрибуция, оригинальный URL подачи заявки и каталог компаний.
Frontend typecheck, lint и build прошли. Существующая схема и Go pgx/sqlc
сохранены; новый ORM или framework не добавлялся.

Автоимпорты и Telegram остаются выключенными. Для следующего ручного импорта
использовать существующий `import-kz-company-jobs` и отдельно подготовленный
allowlist разрешённых источников, как описано в `docs/kz-company-imports.md`.
Перед `--once` обязательно `--dry-run --compare-db`; явно указать Neon
hostname и DB user в защитных переменных импортера. Не запускать старые
GitHub workflows с Supabase Secret против Neon без проверки их guard-условий.

## Подготовлено

- Проект Neon: `floral-frost-85848358`.
- Изолированная ветка: `jobhub-recovery-test` (`br-silent-frog-b533ot8t`),
  родитель — `production`. Срок действия до 16 октября 2026, 13:00 UTC.
- На тестовой ветке применены существующие миграции 1–15;
  итоговая версия 15, `dirty=false`.
- Из сохранённого публичного снимка перенесены 45 тестовых вакансий:
  Kcell — 33, Air Astana — 12. Это не полный дамп Supabase.
- В тесте используются источники `recovery:kcell` и `recovery:airastana`;
  внешняя идентичность — исходный URL. Эти идентификаторы нельзя автоматически
  считать исходными идентификаторами Production. Новые UUID вакансий также
  не заменяют старые связи с откликами или избранным.
- Повторная загрузка обновила 45 записей, не создала дубликатов и не
  деактивировала вакансии. Проверены фильтр Казахстана, чтение карточек,
  ссылки на оригиналы и привязка Kcell к компании с 33 вакансиями.
- Компании не получили проверку автоматически. Записей публикации Telegram — 0.
- На тестовой ветке прошёл существующий `TestAuthServiceIntegration`
  (регистрация, вход и сессии с тестовым email verifier). Реальные email/Google
  и перенос Supabase Auth этим тестом не проверяются.
- Прошли проверки Go для `cmd/migrate`, `internal/database`, `internal/jobs`,
  `internal/ingestion`, `internal/kzcompanies` и `git diff --check`.

## Ограничения переноса старых данных

Production Supabase `rpvwnxsyrtoznnmsyyeg` при проверке вернул
`ECONNREFUSED` на минимальный SQL-запрос. В каталоге проекта и соседних
каталогах ChatGPT не найден дамп базы; `.env.local.backup` содержит настройки,
а не резервную копию PostgreSQL.

Для полного переноса требуется доступ к старой базе или её дамп. После
получения копии нужно восстановить данные на отдельной ветке, сверить ID,
количество вакансий, пользователей, компаний, откликов и избранного.
Сохранять существующие идентификаторы и зависимости; публичный снимок
нельзя использовать для восстановления приватных данных.

Backend PostgreSQL, подтверждение email, Google OAuth и новые резюме
переключены на Neon. `@supabase/supabase-js` удалён. Frontend использует
официальный `@neondatabase/auth`; Go проверяет EdDSA-подпись, issuer,
audience, срок JWT, подтверждённый email и блокировку пользователя.
Google дополнительно требует подтверждённый связанный Google account в
`neon_auth`. Существующие JobHub роли, пароли bcrypt, opaque-сессии и CSRF
сохранены. Это не перенос старых Supabase Auth пользователей.

PDF резюме хранятся в приватном bucket `uploads`, с server-side SigV4 и
path-style S3. Ключи S3 никогда не выдаются браузеру; доступ к файлу проходит
существующие проверки владельца/работодателя. Старые файлы не скопированы.

Старые пользователи, отклики, избранное, сессии и файлы резюме не перенесены.
Не считать их удалёнными или восстановленными: они остаются в недоступном
Supabase. Не запускать автоматические импорты или Telegram при восстановлении.
Supabase не изменялся. Первый deploy сделан без commit/push из
существующего изолированного checkout с сохранённой аварийной опцией.

## Настройки и разработка после перехода

- В production настроены собственный Google OAuth provider и Custom SMTP.
  Redirect origin: `https://jobhub-lovat.vercel.app`; localhost в production
  отключён; email verification
  обязательна. Доставку письма и полный браузерный вход подтвердить вручную.
- Секреты хранятся в Neon/Vercel, не в исходниках. `VITE_NEON_AUTH_URL` —
  публичный Auth URL; `NEON_AUTH_BASE_URL` — тот же URL на сервере.
- Go storage использует `AWS_ENDPOINT_URL_S3`, `AWS_REGION`,
  `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, bucket `uploads` (private).
- `make dev` / `make backend-dev` теперь используют `scripts/dev-neon.sh`.
  Скрипт не применяет миграции и не запускает импорт. Локальная тестовая ветка
  истекает 16 октября 2026, 13:00 UTC; до этой даты подготовить постоянную
  development-ветку и обновить локальный `.env`.
- Supabase session-pooler helper и его опция из `check-public-job` удалены.
  Supabase-specific workflow `apply-telegram-import-guard.yml` удалён из
  репозитория вместе с кодом миграции. Старые SQL-инструкции сохраняются
  для истории инцидента; не применять их к Neon.

Проверки перехода: `go test ./...`, frontend build/typecheck/lint,
`npm audit` (0 vulnerabilities), `git diff --check`, `sh -n scripts/dev-neon.sh`.
На изолированной Neon-ветке прошли регистрация/сессии/CSRF/logout и реальный
S3 round trip PDF, защита от перезаписи и удаление тестового файла. Fixture-тесты
JWT отклоняют чужие issuer/audience, неподтверждённый email, просроченные и
поддельные токены, неверный алгоритм, блокировку и отсутствие subject/expiry.

Старый Supabase проект не удалён: приватные данные ещё не доступны для переноса.
Его удаление до получения копии может уничтожить единственный экземпляр
аккаунтов, откликов и резюме.
