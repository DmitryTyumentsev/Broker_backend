-- +goose Up
create table if not exists refresh_sessions(
    session_id bigserial primary key,
    refresh_token_hash text not null unique,
    device_id text not null,
    created_at timestamptz not null default now(),
    expires_at timestamptz not null,
    revoked_at timestamptz,
    replaced_at timestamptz,
    user_id uuid not null references users(id) on delete cascade,
    constraint unique_refresh_sessions_refresh_token_hash unique(refresh_token_hash) --название верное? сначала тип правила(unique/check/..), затем таблица поле?
);
create index idx_refresh_sessions_user_id on refresh_sessions(user_id);
create index idx_refresh_session_expires_at on refresh_sessions(expires_at);
create index idx_refresh_sessions_refresh_token_hash on refresh_sessions(refresh_token_hash);

-- +goose Down
drop table if exists refresh_sessions;