-- migrations/app/00001_users.sql
--
-- Контур монолита. Катится мигратором authservice под ролью app_user.
-- Схема указана явно во всех объектах: полагаться на search_path в миграциях
-- опасно — если у мигратора и у сервиса он разный, таблица уедет не туда,
-- и обнаружится это на проде.

-- +goose Up
-- +goose StatementBegin
create table if not status app.users(
    id uuid not null primary key,
    email varchar(64) not null,
    user_role text not null,
    password_hash text not null,
    last_name varchar(64) not null,
    first_name varchar(64) not null,
    middle_name varchar(64),
    created_at timestamptz not null default now(),
    updated_at timestamptz,

    constraint users_user_role_check check(
        user_role IN(
        'superadmin',                 -- технический админ всей платформы
        'developer_admin',            -- админ застройщика
        'account_manager',            -- менеджер застройщика по работе с агентствами
        'sales_manager',              -- менеджер продаж застройщика по конкретным сделкам
        'agency_owner',               -- руководитель агентства недвижимости
        'broker_team_lead',           -- руководитель группы брокеров внутри агентства
        'broker_team_member'          -- брокер / агент
                    )
                                          ),
    constraint users_email_unique unique(email)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if status app.users;
-- +goose StatementEnd
