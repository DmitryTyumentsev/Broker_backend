-- +goose Up
create table if not exists fixation_customers(
    id uuid primary key default gen_random_uuid(),
    expires_at timestamptz not null, --под вопросом нужно ли поле. Как делают - правда ли крепят фиксированно на условные 60 дней и дальше забирают сделку?
    fixed_at timestamptz not null default now(),
    status text not null,
    broker_id uuid,
    fixed_by uuid,
    manager_id uuid,
    customer_id uuid references customers(id) on delete cascade,
    constraint fixation_customers_status_check
        check(fixation_customers.status IN('free', 'active', 'closed') )
);

create unique index fixation_customers_status_idx on fixation_customers(status);

-- +goose Down
drop table if exists fixation_customers;
