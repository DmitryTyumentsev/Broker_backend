-- +goose Up
create table if not exists fixations(
    id uuid primary key default gen_random_uuid(),
    fixed_at timestamptz not null default now(),
    expires_at timestamptz,
    status text not null,
    broker_id uuid not null,
    fixed_by uuid not null,
    fix_for uuid not null,
    project_id uuid not null, --брокер крепит за собой клиента по его номеру телефона и указывает объект. другой брокер не может закрепить этого клиента на этом объекте, но может на другом объекте
    phone_hash text not null,
    constraint fixations_status_check
        check(status IN('active', 'converted', 'expired', 'removed') )
);
create partial unique index if not exists fixations_phone_hash_unique on fixations(phone_hash); --почему partial красным горит? как вообще все работает: каким образом индексы лочат запросы параллельные? что вообще отвечает за блокировку и кто кроме индексов может блокировать? нет картины у меня единой тут сейчас

-- +goose Down
drop table if exists fixations;
