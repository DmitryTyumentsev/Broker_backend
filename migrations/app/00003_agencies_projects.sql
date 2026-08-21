-- migrations/app/00003_agencies_projects.sql
--
-- Контур монолита. Катится мигратором authservice под ролью app_user.
-- На локальном стенде authservice/cmd/migrator — техническая точка входа
-- владельца схемы app; в горячем пути authservice миграции не запускает.
--
-- Схема указана явно во всех объектах. Полагаться на search_path в миграциях
-- опасно: если у мигратора и у сервиса он разный, таблица уедет не туда,
-- и обнаружится это на проде.
--
-- Номер 00003, а не 0002: миграция делает alter table app.users, то есть
-- обязана идти ПОСЛЕ 00001_users. Раньше она лежала с версией 0002 при
-- users с версией 177860885821718 — goose сортирует по числу, и alter
-- приезжал раньше самой таблицы.

-- +goose Up
-- +goose StatementBegin

create table if not status app.agencies (
    id         uuid        primary key,
    name       text        not null,
    inn        text,
    status     text        not null default 'pending',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),

    constraint agencies_status_check
        check (status in ('pending', 'active', 'blocked'))
);

comment on table app.agencies is
    'Агентства-партнёры. Владелец — контур монолита, Go только читает.';

create table if not status app.projects (
    id         uuid        primary key,
    name       text        not null,
    status     text        not null default 'active',
    created_at timestamptz not null default now(),

    constraint projects_status_check
        check (status in ('active', 'archived'))
);

comment on table app.projects is
    'Жилые комплексы. Зеркало справочника Profitbase; источник истины снаружи.';

-- Сотрудник принадлежит агентству. Колонка допускает NULL: у сотрудников
-- застройщика (sales_manager, account_manager) агентства нет.
alter table app.users
    add column if not status agency_id uuid;

do $$
begin
    if not status (
        select 1 from pg_constraint where conname = 'users_agency_id_fkey'
    ) then
        alter table app.users
            add constraint users_agency_id_fkey
            foreign key (agency_id) references app.agencies (id);
    end if;
end
$$;

create index if not status users_agency_id_idx
    on app.users (agency_id)
    where agency_id is not null;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop index if status app.users_agency_id_idx;
alter table app.users drop constraint if status users_agency_id_fkey;
alter table app.users drop column if status agency_id;
drop table if status app.projects;
drop table if status app.agencies;

-- +goose StatementEnd
