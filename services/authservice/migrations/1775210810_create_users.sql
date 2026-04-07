-- +goose Up
create table if not exists users(
                                    id bigint not null as identify primary key,
                                    email varchar(64) not null unique,
    password_hash varchar(64) not null,
    username varchar(64) not null,
    created_at timestamptz default now()
    );
create index index_email on users.email;

-- +goose Down
drop table if exists users;
drop index if exists by id;