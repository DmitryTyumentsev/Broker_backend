-- migrations/integration/00001_fixations.sql
--
-- Контур Go. Катится мигратором fixationservice под ролью go_user.
-- Схема указана явно: таблица обязана лежать в integration, а не там,
-- куда покажет search_path конкретного соединения.

-- +goose Up
-- +goose StatementBegin
create table if not exists integration.fixations(
    id uuid primary key default gen_random_uuid(),
    fixed_at timestamptz not null default now(),
    expires_at timestamptz not null,
    status text not null,
    agency_id uuid not null,
    fix_by uuid not null,
    fix_for uuid not null,
    project_id uuid not null,
    phone_hash text not null,
    constraint fixations_status_check
        check(status IN('active', 'converted', 'expired', 'removed') )
);
create unique index if not exists fixations_phone_hash_project_id_idx on integration.fixations(phone_hash, project_id) where status = 'active';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists integration.fixations;
-- +goose StatementEnd
