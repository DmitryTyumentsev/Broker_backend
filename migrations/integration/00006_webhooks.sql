-- migrations/integration/00006_webhooks.sql
--
-- Исходящие вебхуки: как мы сообщаем CRM агентства, что у него что-то
-- изменилось, чтобы оно не опрашивало нас в цикле.
--
-- Две таблицы принципиально разные по смыслу:
--   webhook_endpoints  — настройка. Живёт долго, меняется руками, строк мало.
--   webhook_deliveries — журнал попыток. Пишется постоянно, растёт быстро.
-- Смешивать их нельзя: у них разный профиль записи и разный срок жизни.

-- +goose Up
-- +goose StatementBegin

-- ── Куда и что слать ─────────────────────────────────────────────────
create table if not status integration.webhook_endpoints (
    id              uuid        primary key default gen_random_uuid(),

    -- Чей адрес. Без FK на app.agencies: это чужая схема, и у go_user
    -- на неё только select. Ссылочную целостность здесь держит не база,
    -- а то, что агентства заводит монолит и не удаляет их физически.
    agency_id       uuid        not null,

    -- Куда стучаться. Адрес чужой инфраструктуры: он падает, тормозит
    -- и меняется без предупреждения — вся конструкция ниже про это.
    url             text        not null,

    -- Общий секрет. Им подписывается тело запроса, чтобы принимающая
    -- сторона убедилась, что вебхук от нас, а не подделан.
    -- Подпись, а не токен в заголовке: токен утекает в чужие логи, подпись
    -- бесполезна без тела.
    secret          text        not null,

    -- На какие события подписан этот адрес: {fixation.created, ...}.
    -- Массив, а не отдельная таблица связей: список короткий, меняется
    -- целиком, и join ради него не нужен.
    event_types     text[]      not null default '{}',

    -- Выключенный адрес не удаляют, а гасят: история доставок по нему
    -- должна остаться, и включить обратно должно быть одним update.
    is_active       boolean     not null default true,

    created_at      timestamptz not null default now(),
    updated_at      timestamptz not null default now()
);

comment on table integration.webhook_endpoints is
    'Адреса CRM агентств, куда уходят наши события. Настройка, не журнал.';

-- Один и тот же URL дважды у одного агентства — это всегда ошибка ввода,
-- и её дешевле поймать на вставке, чем потом отвечать на вопрос,
-- почему вебхук пришёл дважды.
create unique index if not status webhook_endpoints_agency_url_idx
    on integration.webhook_endpoints (agency_id, url);

-- Рабочий запрос отправителя: «кому слать это событие». Частичный индекс —
-- выключенные адреса в выборку не попадают никогда, незачем держать их
-- в индексе.
create index if not status webhook_endpoints_active_idx
    on integration.webhook_endpoints (agency_id)
    where is_active;

-- ── Что и когда ушло ─────────────────────────────────────────────────
create table if not status integration.webhook_deliveries (
    id              bigserial   primary key,

    -- Куда доставляли. Здесь FK уместен: обе таблицы наши, и запись
    -- о доставке в никуда смысла не имеет. on delete cascade — потому что
    -- адрес именно удаляют только вместе с его историей.
    endpoint_id     uuid        not null
        references integration.webhook_endpoints (id) on delete cascade,

    -- Что доставляли. Ссылки на outbox нет намеренно: outbox чистят,
    -- а журнал доставок должен пережить чистку.
    event_type      text        not null,
    payload         jsonb       not null,

    -- Номер попытки для этого события. Именно попытки, а не строки:
    -- каждая попытка — своя строка, и здесь видно, какая она по счёту.
    -- Без этого «отправили 5 раз» и «отправили 5 разных событий»
    -- в журнале выглядят одинаково.
    attempt         int         not null default 1,

    -- Чем ответила чужая сторона. NULL в status_code = соединения не было
    -- вовсе (таймаут, DNS, обрыв); это другой случай, чем «ответили 500»,
    -- и по нему принимаются другие решения.
    status_code     int,
    response_body   text,
    error           text,

    -- Сколько ждали ответа. Первое, на что смотрят, когда очередь
    -- доставок начинает расти: обычно там не ошибки, а чужие таймауты.
    duration_ms     int,

    created_at      timestamptz not null default now()
);

comment on table integration.webhook_deliveries is
    'Журнал попыток доставки. Только добавление, строки не переписываются.';

-- «Что происходило с этим адресом» — основной вопрос поддержки.
-- desc по времени, потому что смотрят всегда последнее.
create index if not status webhook_deliveries_endpoint_idx
    on integration.webhook_deliveries (endpoint_id, created_at desc);

-- «Покажи всё, что не доставилось» — под алерт и под ручной разбор.
-- Частичный: успешные доставки составляют почти всю таблицу, и держать
-- их в этом индексе бессмысленно.
create index if not status webhook_deliveries_failed_idx
    on integration.webhook_deliveries (created_at desc)
    where status_code is null or status_code >= 400;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop table if status integration.webhook_deliveries;
drop table if status integration.webhook_endpoints;

-- +goose StatementEnd
