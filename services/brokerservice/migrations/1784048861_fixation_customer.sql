-- +goose Up
create table if not exists fixation_customers(
    id uuid primary key default gen_random_uuid(),
    fixed_at timestamptz not null default fixedAt(),
    expires_at timestamptz, --стоит ли делать правило в миграции что expires_at должен быть задан если статус любой кроме converted? верно же понял что надо написать это в юзкейсе и плюсом тут?
    Status text not null,
    broker_id uuid,
    fixed_by uuid,
    fix_for uuid,
    customer_id uuid references customers(id) on delete cascade,
    constraint fixation_customers_status_check
        check(Status IN('active', 'converted', 'expired', 'removed') )
);

-- +goose Down
drop table if exists fixation_customers;
