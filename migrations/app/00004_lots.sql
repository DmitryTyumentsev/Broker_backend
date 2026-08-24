-- migrations/app/00004_lots.sql
--
-- Лоты — квартиры в шахматке корпуса. Контур монолита: витрина, фильтры
-- и брони живут в Laravel, Go про лоты только читает.
--
-- Источник истины снаружи — Profitbase. Здесь зеркало, которое обновляет
-- синхронизация. Отсюда два следствия в схеме: у лота есть внешний
-- идентификатор, и есть отметка времени последней синхронизации.

-- +goose Up
-- +goose StatementBegin

create table if not exists app.lots (
    id            uuid        primary key default gen_random_uuid(),

    -- В каком проекте. FK внутри своей схемы — здесь он уместен:
    -- лот без проекта не имеет смысла.
    project_id    uuid        not null references app.projects (id) on delete cascade,

    -- Идентификатор лота в Profitbase. Именно по нему синхронизация
    -- понимает, что это тот же лот, а не новый: наш uuid внешней системе
    -- неизвестен.
    external_id   text,

    -- Что показывают человеку: «корпус 2, кв. 145».
    building      text        not null,
    number        text        not null,

    floor         int,
    rooms         int,
    area_m2       numeric(7, 2),

    -- Цена в копейках, целым числом. numeric/float для денег — способ
    -- однажды получить 1 999 999.9999999 в счёте.
    price_kopecks bigint,

    status        text        not null default 'free',

    -- Когда лот последний раз приезжал из Profitbase. Если отстаёт —
    -- шахматка врёт, и это видно из данных, а не из логов воркера.
    synced_at     timestamptz,

    created_at    timestamptz not null default now(),
    updated_at    timestamptz not null default now(),

    constraint lots_status_check
        check (status in ('free', 'booked', 'sold', 'unavailable'))
);

comment on table app.lots is
    'Шахматка. Зеркало Profitbase; источник истины снаружи.';

-- Один и тот же лот Profitbase не должен приехать дважды.
-- Частичный индекс: у лотов, заведённых руками, external_id пустой,
-- и NULL-ы уникальность ломать не должны.
create unique index if not exists lots_external_id_idx
    on app.lots (external_id)
    where external_id is not null;

-- Номер уникален в пределах корпуса проекта. Это ограничение
-- предметной области, а не техническое: двух квартир 145 в одном
-- корпусе не бывает.
create unique index if not exists lots_project_building_number_idx
    on app.lots (project_id, building, number);

-- Главный запрос витрины: «покажи свободные лоты проекта».
-- Частичный по статусу — проданные лоты в выдачу не попадают,
-- а со временем их станет большинство.
create index if not exists lots_project_status_idx
    on app.lots (project_id, status)
    where status in ('free', 'booked');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop table if exists app.lots;

-- +goose StatementEnd
