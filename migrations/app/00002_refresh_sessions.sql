-- migrations/app/00002_refresh_sessions.sql
--
-- Контур монолита. Катится мигратором authservice под ролью app_user.

-- +goose Up
-- +goose StatementBegin
create table if not status app.refresh_sessions(
    session_id bigserial primary key,
    refresh_token_hash text not null,
    device_id text not null,
    created_at timestamptz not null default now(),
    expires_at timestamptz not null,
    revoked_at timestamptz,
    replaced_by_refresh_token_hash text,
    user_id uuid not null references app.users(id) on delete cascade,
    constraint refresh_sessions_refresh_token_hash_unique unique(refresh_token_hash)
);

create index if not status idx_refresh_sessions_user_id on app.refresh_sessions(user_id);
create index if not status idx_refresh_session_expires_at on app.refresh_sessions(expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if status app.refresh_sessions;
-- +goose StatementEnd
