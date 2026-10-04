# Исследование источников вакансий компаний Казахстана — первая партия

Дата: 2026-09-25. Проверены 20 явно разрешённых URL в режиме только чтения.
База, импорт, staging, Production и реестр источников не изменялись.

## Метод и ограничения

Каждый URL обрабатывался существующим безопасным Go-сканером: до 15 HTTP-запросов
на сайт, проверка SSRF/публичного адреса, ответ до 3 МиБ. Сайты проверялись
последовательно. Всего **69 запросов из максимальных 300**. Учётные данные,
cookie, тела запросов и личные данные не собирались.

Сначала все 20 сайтов прошли статическую проверку. Браузерный fallback в этой
партии **не запускался**: у renderer внутренний бюджет 100 запросов, а
исследование требовало отдельного ограничения 40 на сайт. Код ради
неограниченной диагностики не менялся. Нулевой статический результат означает
«не выяснено» или «заблокировано», а не отсутствие вакансий на сайте.

## Результаты

| Компания | URL | Классификация | Статических запросов | URL вакансий | Нормализовано | Пропущено | Состояние / причина |
|---|---|---:|---:|---:|---:|---:|---|
| Halyk Bank | https://halykbank.kz/about/career | STATIC_HTML | 4 | 2 | 2 | 0 | успешно |
| Tele2 | https://tele2.kz/career | JS_SPA_UNRESOLVED | 2 | 0 | 0 | 0 | успешно; статических ссылок на вакансии нет |
| Technodom | https://hr.technodom.kz/ | STATIC_HTML | 15 | 39 | 12 | 27 | исчерпан лимит запросов |
| Sulpak | https://job.sulpak.kz/WorkMap/WorkMap | JS_SPA_UNRESOLVED | 2 | 0 | 0 | 0 | успешно; статических ссылок на вакансии нет |
| BI Group | https://bi.group/ru/career | BLOCKED | 1 | 0 | 0 | 0 | HTTP_FAILED |
| Kolesa Group | https://kolesa.group/career/job | STATIC_HTML | 5 | 3 | 3 | 0 | успешно |
| ERG | https://erg.kz/ru/career | JS_SPA_UNRESOLVED | 2 | 0 | 0 | 0 | успешно; статических ссылок на вакансии нет |
| Magnum | https://hr.magnum.kz/ | JS_SPA_UNRESOLVED | 3 | 0 | 0 | 0 | успешно; статических ссылок на вакансии нет |
| Kazakhmys | https://www.kazakhmys.kz/ru/career | JS_SPA_UNRESOLVED | 2 | 0 | 0 | 0 | успешно; статических ссылок на вакансии нет |
| Qarmet | https://job.qarmet.kz/ | BLOCKED | 1 | 0 | 0 | 0 | TRANSPORT_FAILED |
| QazaqGaz | https://www.qazaqgaz.kz/ru/karera | JS_SPA_UNRESOLVED | 2 | 0 | 0 | 0 | успешно; статических ссылок на вакансии нет |
| Samruk-Energy | https://samruk-energy.kz/ru/company/vacancy | JS_SPA_UNRESOLVED | 2 | 0 | 0 | 0 | успешно; статических ссылок на вакансии нет |
| Kazatomprom | https://hrekap.kazatomprom.kz/ | JS_SPA_UNRESOLVED | 3 | 0 | 0 | 0 | успешно; статических ссылок на вакансии нет |
| KazMunayGas | http://work.kmg.kz/ | JS_SPA_UNRESOLVED | 4 | 0 | 0 | 0 | перенаправление на HTTPS; статических ссылок на вакансии нет |
| Tengizchevroil | https://tcoext.azure.chevron.com/ru/career/vacancy | BLOCKED | 1 | 0 | 0 | 0 | TRANSPORT_FAILED |
| SMALL | https://small.kz/ru/astana/job | BLOCKED | 1 | 0 | 0 | 0 | HTTP_FAILED |
| KEGOC | https://www.kegoc.kz/ | JS_SPA_UNRESOLVED | 2 | 0 | 0 | 0 | URL раздела карьеры не найден в статическом HTML |
| RG Brands | https://rgbrands.com/career | JS_SPA_UNRESOLVED | 3 | 0 | 0 | 0 | успешно; статических ссылок на вакансии нет |
| Beeline Kazakhstan | https://beeline.kz/ru | JS_SPA_UNRESOLVED | 13 | 0 | 0 | 0 | успешно; статических ссылок на вакансии нет |
| Astana Motors Manufacturing Kazakhstan | https://www.astanamotors-manufacturing.kz/ | BLOCKED | 1 | 0 | 0 | 0 | HTTP_FAILED |

Классификация намеренно осторожная. `JS_SPA_UNRESOLVED` означает, что
ограниченная статическая проверка не нашла пригодных записей; наличие
браузерного/API endpoint не доказано. Известная ATS среди подтверждённых
источников на этом этапе не обнаружена.

## Примеры нормализованных записей

Три успешных статических источника дали следующие публичные URL:

- Halyk Bank: [vacancies/2](https://halykbank.kz/about/career/vacancies/2), [vacancies/1](https://halykbank.kz/about/career/vacancies/1). Страница представляет список вакансий. Сканер сохранил исходный URL и не выдумывал работодателя, место, зарплату и занятость там, где страница не давала этих полей.
- Technodom: [Оператор котельной](https://hr.technodom.kz/vacancies/operator-kotelnoj/), [Оператор ПРТ](https://hr.technodom.kz/vacancies/operator-prt/), [Энергетик](https://hr.technodom.kz/vacancies/energetik/), [Инженер-электрик](https://hr.technodom.kz/vacancies/inzhener-elektrik-2/), [Слесарь-сантехник](https://hr.technodom.kz/vacancies/slesar-santehnik/). Нормализовано 12 записей; 27 ссылок пропущены как некорректные или не ведущие на вакансии в пределах бюджета 15 запросов.
- Kolesa Group: [Middle GO Backend-разработчик Core team](https://kolesa.group/career/job/middle-go-backend-razrabotcik-core-team-136951220), [Fullstack Web QA](https://kolesa.group/career/job/fullstack-web-qa-v-krishakz-136312001), [Разработчик бизнес процессов (Middle)](https://kolesa.group/career/job/razrabotcik-biznes-processov-middle-137563145). Нормализованы три записи.

Нормализованные записи не сохранялись. Браузерных примеров на этом этапе нет.

## Границы следующего этапа

Неопределённые сайты можно проверить отдельно разрешённой браузерной/API
диагностикой: не более пяти сайтов и 40 запросов на каждый. Этот документ
не разрешает такую проверку, создание адаптера, запись в реестр или
включение в Production. Ранее проверенные Kaspi, Kcell и Air Astana
повторно не сканировались.

## Дополнительная браузерная проверка

Renderer теперь принимает меньший бюджет от вызывающего кода: фактически
`min(requested, 100)`; нулевое и отрицательное значения оставляют лимит 100.
Выбраны пять доступных специализированных страниц карьеры с неопределённым
результатом статической проверки. API/ATS-адаптеры не создавались.

| Компания | Браузер | Лимит | Запросов | URL вакансий | API/ATS | Стабильный ID | Найдено | Классификация | Следующий шаг |
|---|---|---:|---:|---:|---|---|---:|---|---|
| Tele2 | резервный Chromium | 40 | 0 | 0 | не обнаружен | неизвестен | неизвестно | JS_SPA_UNRESOLVED | отдельно согласовать диагностику публичного API/сетевых запросов |
| Sulpak | резервный Chromium | 40 | 28 | 0 | не обнаружен | неизвестен | 0 наблюдаемых | JS_SPA_UNRESOLVED | проверить согласованный публичный endpoint, если он задокументирован |
| ERG | резервный Chromium | 40 | 0 | 0 | не обнаружен | неизвестен | неизвестно | JS_SPA_UNRESOLVED | отдельно согласовать диагностику публичного API/сетевых запросов |
| Magnum | резервный Chromium | 40 | 0 | 0 | не обнаружен | неизвестен | неизвестно | JS_SPA_UNRESOLVED | отдельно согласовать диагностику публичного API/сетевых запросов |
| Kazakhmys | резервный Chromium | 40 | 40 | 0 | не обнаружен | неизвестен | 0 наблюдаемых | JS_SPA_UNRESOLVED | изучать только документированный публичный API или ленту сайта |

Для Tele2, ERG и Magnum статическая проверка не определила подходящую SPA,
поэтому fallback не сработал. Sulpak и Kazakhmys отобразились в пределах
40 запросов, но не дали нормализованных вакансий; Kazakhmys исчерпал лимит.
Публичных JSON-записей, стабильных ID, контракта пагинации и ATS endpoint
не найдено. «0 наблюдаемых» означает отсутствие записей в обработанном
DOM/structured-data, а не отсутствие вакансий у компании.
