-- Схемы и роли двух контуров. Катится ОДИН РАЗ суперпользователем,
-- до всех остальных миграций:
--   psql -U postgres -d fixation -f migrations/bootstrap/0001_schemas_and_roles.sql

-- +goose Up
-- +goose StatementBegin
create schema if not exists app;           -- владелец: монолит (Laravel / authservice)
create schema if not exists integration;   -- владелец: Go-контур

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

grant usage on schema app, integration to app_user, go_user;

-- монолит: полные права на свою схему, только чтение чужой
grant all privileges on all tables in schema app         to app_user;
grant select         on all tables in schema integration to app_user;

-- Go: зеркально
grant all privileges on all tables in schema integration to go_user;
grant select         on all tables in schema app         to go_user;

-- то же для таблиц, которые появятся позже
alter default privileges in schema app         grant all    on tables to app_user;
alter default privileges in schema app         grant select on tables to go_user;
alter default privileges in schema integration grant all    on tables to go_user;
alter default privileges in schema integration grant select on tables to app_user;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop schema if exists integration cascade;
drop schema if exists app cascade;
-- +goose StatementEnd
