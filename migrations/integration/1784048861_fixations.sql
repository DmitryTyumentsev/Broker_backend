-- +goose Up
create table if not exists fixations(
    id uuid primary key default gen_random_uuid(), --как лучше - оставить id или сделать составной ключ из phone_hash и project_id? когда как лучше делать? какой синтаксис для составного ключа?
    fixed_at timestamptz not null default now(),
    expires_at timestamptz,
    status text not null,
    broker_id uuid not null,
    fixed_by uuid not null,
    fix_for uuid not null,
    project_id uuid not null,
    phone_hash text not null,
    constraint fixations_status_check
        check(status IN('active', 'converted', 'expired', 'removed') )
);
create unique index if not exists fixations_phone_hash_unique on fixations(phone_hash)
where ;

-- +goose Down
drop table if exists fixations;
