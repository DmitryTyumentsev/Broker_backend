-- migrations/bootstrap/00001_schemas_roles_extensions.sql
--
-- Каталог bootstrap катится ОДИН РАЗ и ТОЛЬКО суперпользователем, до всех
-- остальных миграций. Здесь лежит ровно то, на что у app_user и go_user
-- прав нет и не будет: схемы, роли, гранты, расширения.
--
--   make migrate-bootstrap
--   или вручную:
--   psql -U postgres -d broker -f migrations/bootstrap/00001_schemas_roles_extensions.sql
--
-- Дальше каждый контур катит свой каталог своей ролью:
--   migrations/app         → authservice/cmd/migrator, роль app_user
--   migrations/integration → fixationservice/cmd/migrator, роль go_user

-- +goose Up
-- +goose StatementBegin

-- ── Расширения ────────────────────────────────────────────────────────
-- pgcrypto нужен ради gen_random_uuid() в default-значениях.
-- В PostgreSQL 13 gen_random_uuid() уже есть в ядре, но расширение ставим
-- явно: на 11/12, откуда мигрируют стенды, его в ядре нет, и миграция,
-- которая работает только на 13, — мина под первый же старый инстанс.
create extension if not exists pgcrypto;

-- ── Схемы ─────────────────────────────────────────────────────────────
create schema if not exists app;           -- владелец: монолит (Laravel / authservice)
create schema if not exists integration;   -- владелец: Go-контур

-- ── Роли ──────────────────────────────────────────────────────────────
-- Пароли здесь локальные и стендовые. В проде роли заводит DBA, а пароли
-- приезжают из секрет-хранилища — эта миграция там не катится вообще.
do $$
begin
  if not exists (select 1 from pg_roles where rolname = 'app_user') then
    create role app_user login password 'app_user';
  end if;
  if not exists (select 1 from pg_roles where rolname = 'go_user') then
    create role go_user login password 'go_user';
  end if;
end
$$;

-- ── Гранты ────────────────────────────────────────────────────────────
-- Это и есть граница контуров. Не соглашение в вики, а отказ базы:
-- если Go полезет писать в app.*, он получит permission denied,
-- а не тихо испортит чужие данные.
grant usage on schema app, integration to app_user, go_user;

-- монолит: полные права на свою схему, только чтение чужой
grant all privileges on all tables in schema app         to app_user;
grant select         on all tables in schema integration to app_user;

-- Go: зеркально
grant all privileges on all tables in schema integration to go_user;
grant select         on all tables in schema app         to go_user;

-- Последовательности: bigserial создаёт отдельный объект, и права на него
-- выдаются отдельно. Без этого insert упадёт на nextval().
grant usage, select on all sequences in schema app         to app_user;
grant usage, select on all sequences in schema integration to go_user;

-- То же самое для объектов, которые появятся позже. Без default privileges
-- каждая новая таблица приезжала бы без прав, и падало бы через неделю
-- после миграции — в момент, когда её уже никто не помнит.
alter default privileges in schema app         grant all    on tables      to app_user;
alter default privileges in schema app         grant select on tables      to go_user;
alter default privileges in schema integration grant all    on tables      to go_user;
alter default privileges in schema integration grant select on tables      to app_user;

alter default privileges in schema app         grant usage, select on sequences to app_user;
alter default privileges in schema integration grant usage, select on sequences to go_user;

-- Мигратор каждого контура пишет свою таблицу версий в свою схему
-- (app.goose_db_version / integration.goose_db_version). Права на создание
-- таблиц в своей схеме нужны именно для этого.
grant create on schema app         to app_user;
grant create on schema integration to go_user;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop schema if exists integration cascade;
drop schema if exists app cascade;

-- +goose StatementEnd
