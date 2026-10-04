---
title: "JobHub Docs"
description: "Интерактивная документация JobHub для новых разработчиков."
---

<img src="/docs/assets/jobhub-logo.png" alt="Логотип JobHub" width="260" />

# JobHub Docs

Добро пожаловать в документацию JobHub. Этот сайт помогает новым
разработчикам быстро понять, как устроен проект: страницы, пользовательские
сценарии, frontend, backend, API, Supabase, деплой и Telegram publisher.

## С чего начать

- [О проекте](/docs/project-overview) — задача, роли, технологии и структура
  репозитория.
- [Инструкция для нового разработчика](/docs/new-developer-guide) — первый
  запуск, переменные окружения и безопасная работа.
- [Архитектура системы](/docs/system-architecture) — потоки
  frontend → backend → PostgreSQL/Supabase и Telegram publisher.
- [Frontend](/docs/frontend-pages-reference) — маршруты, страницы, кнопки,
  действия и API-вызовы.
- [Backend и API](/docs/backend-api-reference) — endpoints, auth, ошибки и
  исходный код.

## Основные разделы

- **Приложение** — сценарии кандидата и работодателя, профили компаний,
  видимость вакансий и резюме.
- **Backend** — API, сессии, миграции, безопасность и операционные заметки.
- **Вакансии** — источники, импорт, правила публикации и качество данных.
- **Telegram** — как новая вакансия попадает в Telegram-канал без дублей.
- **Проверки** — чек-листы ревью, тестирования, релиза и интеграций.

## Что уже проверено по коду

Документация опирается на `frontend/src/App.tsx`, `frontend/src/api`,
`backend/internal/handlers`, `backend/migrations`, `backend/internal/telegrampublisher`,
`vercel.json`, `Makefile` и `.github/workflows/telegram-publisher.yml`.
Неподтверждённые детали помечены как «не подтверждено».

## Production

- Сайт: [jobhub-lovat.vercel.app](https://jobhub-lovat.vercel.app/)
- Репозиторий: [juniorqazaq/JobHub](https://github.com/juniorqazaq/JobHub)
