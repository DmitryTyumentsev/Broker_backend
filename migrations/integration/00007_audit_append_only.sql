-- migrations/integration/00007_audit_append_only.sql
--
-- integration.audit_log создаёт go_user, поэтому он является владельцем
-- таблицы и сохраняет право менять её ACL. REVOKE из 00002 запрещает
-- UPDATE/DELETE сейчас; FORCE ROW LEVEL SECURITY добавляет второй рубеж:
-- даже после ошибочного GRANT для этих операций по-прежнему нет политик.

-- +goose Up
-- +goose StatementBegin

alter table integration.audit_log enable row level security;
alter table integration.audit_log force row level security;

create policy audit_log_read
    on integration.audit_log
    for select
    to go_user, app_user
    using (true);

create policy audit_log_append
    on integration.audit_log
    for insert
    to go_user
    with check (true);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop policy if exists audit_log_append on integration.audit_log;
drop policy if exists audit_log_read on integration.audit_log;
alter table integration.audit_log no force row level security;
alter table integration.audit_log disable row level security;

-- +goose StatementEnd
