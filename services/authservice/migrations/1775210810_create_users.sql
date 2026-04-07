-- +goose Up
create table if not exists users(
    id bigint not null generated always as identity primary key,
    email varchar(64) not null unique,
    password_hash text not null,
    username varchar(64) not null,
    created_at timestamptz default now()
    );

-- +goose Down
drop table if exists users;