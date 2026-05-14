-- +goose Up
create table if not exists users(
    id uuid not null primary key,
    email varchar(64) not null unique,
    user_role text not null,
    password_hash text not null,
    last_name varchar(64) not null,
    first_name varchar(64) not null,
    middle_name varchar(64),
    created_at timestamptz default now(),
    updated_at timestamptz default now(),

    constraint constraint_check_role check(
        user_role IN(
        'superadmin',                 -- технический админ всей платформы
        'developer_admin',            -- админ застройщика
        'account_manager',            -- менеджер застройщика по работе с агентствами
        'sales_manager',              -- менеджер продаж застройщика по конкретным сделкам
        'agency_owner',               -- руководитель агентства недвижимости
        'broker_team_lead',           -- руководитель группы брокеров внутри агентства
        'broker'                     -- брокер / агент
                    )
                                          )
    );
create index idx_users_email on users(email);

-- +goose Down
drop table if exists users;