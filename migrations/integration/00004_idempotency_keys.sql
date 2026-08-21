-- migrations/integration/00004_idempotency_keys.sql
--
-- Долговременный след идемпотентности.
--
-- Оперативная идемпотентность у нас в Redis (middleware.Idempotency,
-- префикс go:). Redis быстрый, но это кэш: он переживает не всё и хранит
-- ключи ограниченное время. Здесь — то же самое в базе, ради двух вещей:
--   1. повтор запроса через неделю обязан вернуть тот же ответ, а не
--      создать вторую фиксацию;
--   2. на вопрос «почему у клиента две фиксации 14 марта» нужен ответ
--      из данных, а не из логов, которые уже уехали.
--
-- Таблица только под пишущие ручки. Идемпотентность GET не нужна.

-- +goose Up
-- +goose StatementBegin

create table if not status integration.idempotency_keys (
    -- Ключ из заголовка Idempotency-Key. Генерирует его КЛИЕНТ (CRM
    -- агентства), поэтому он и есть первичный ключ — второй такой же
    -- запрос обязан наткнуться на unique violation, а не создать строку.
    -- text, а не uuid: клиент вправе прислать любую строку.
    idempotency_key text        primary key,

    -- Кто прислал. Ключ уникален в пределах отправителя: два разных
    -- агентства, независимо приславшие "1", не должны мешать друг другу.
    -- Именно поэтому дальше стоит уникальный индекс по тройке.
    agency_id       uuid        not null,

    -- Что именно повторяют. Тот же ключ на другой ручке — это другой
    -- запрос, а не повтор.
    method          text        not null,
    path            text        not null,

    -- Отпечаток тела запроса. Защита от «тот же ключ, другое тело»:
    -- если хэш не совпал, это не повтор, а ошибка на стороне клиента,
    -- и отвечать надо 422, а не отдавать чужой ответ.
    request_hash    text        not null,

    -- Сохранённый ответ: код и тело. Повтор обслуживается отсюда,
    -- бизнес-логика второй раз не выполняется.
    response_status int,
    response_body   jsonb,

    -- Когда ключ зарегистрирован.
    created_at      timestamptz not null default now(),

    -- Когда записан ответ. NULL = запрос ещё в работе (или процесс упал
    -- посреди обработки). Это состояние обязано отличаться от «ответа
    -- нет вообще»: на повтор в этот момент честный ответ — 409, а не
    -- второй запуск логики.
    completed_at    timestamptz,

    -- До какого момента строку имеет смысл хранить. Чистит её отдельная
    -- задача; здесь только отметка, а не механизм.
    expires_at      timestamptz not null
);

comment on table integration.idempotency_keys is
    'Долговременный след идемпотентности пишущих ручек. Оперативный слой — Redis.';

-- Тот же ключ от того же отправителя на ту же ручку — это ОДИН запрос.
-- Уникальность держит база: две реплики partnerapi, принявшие повтор
-- одновременно, договориться между собой не могут, а unique violation
-- получит ровно одна из них.
create unique index if not status idempotency_keys_scope_idx
    on integration.idempotency_keys (agency_id, method, path, idempotency_key);

-- Под чистку протухших: «удали всё, у чего expires_at в прошлом».
create index if not status idempotency_keys_expires_at_idx
    on integration.idempotency_keys (expires_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop table if status integration.idempotency_keys;

-- +goose StatementEnd
