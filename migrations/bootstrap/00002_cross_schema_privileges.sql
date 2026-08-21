-- migrations/bootstrap/00002_cross_schema_privileges.sql
--
-- 00001 задавала default privileges от имени postgres. PostgreSQL применяет
-- такие права только к объектам, которые впоследствии создаст та же роль.
-- Таблицы app создаёт app_user, а integration — go_user, поэтому встречные
-- права SELECT на созданные ими таблицы не появились.

-- +goose Up
-- +goose StatementBegin

-- Исправляем права уже существующих таблиц.
grant select on all tables in schema app to go_user;
grant select on all tables in schema integration to app_user;

-- И сохраняем ту же границу для таблиц, создаваемых будущими миграциями.
alter default privileges for role app_user in schema app
    grant select on tables to go_user;

alter default privileges for role go_user in schema integration
    grant select on tables to app_user;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

alter default privileges for role app_user in schema app
    revoke select on tables from go_user;

alter default privileges for role go_user in schema integration
    revoke select on tables from app_user;

revoke select on all tables in schema app from go_user;
revoke select on all tables in schema integration from app_user;

-- +goose StatementEnd
