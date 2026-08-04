-- migrations/integration/0002_audit_log.sql
--
-- Журнал аудита. Катится мигратором fixationservice под ролью go_user.
--
-- Главное свойство: только добавление. Оно обеспечивается не соглашением,
-- а отзывом прав в конце файла — приложение физически не может
-- переписать или удалить запись.

-- +goose Up
-- +goose StatementBegin

create table if not exists integration.audit_log (
    id              bigserial   primary key,

    -- на что смотрим. Внешнего ключа НЕТ намеренно: аудит обязан пережить
    -- удаление объекта, иначе он перестаёт быть аудитом
    entity_type     text        not null,
    entity_id       uuid        not null,

    -- что произошло. Это событие, а не статус: 'created' — не то же самое,
    -- что status = 'active'
    action          text        not null,

    -- состояние до и после. При создании before пустой.
    -- Именно эти две колонки делают запись самодостаточной:
    -- ответ на «что тут было» не требует обращения к текущему состоянию,
    -- которое как раз и оспаривают
    state_before    jsonb,
    state_after     jsonb,

    -- кто. Только из принципала, никогда из тела запроса
    actor_user_id   uuid,
    actor_agency_id uuid,
    actor_role      text,

    -- чем связать с логами и как поймать перебор телефонов
    request_id      text,
    client_ip       inet,

    created_at      timestamptz not null default now(),

    constraint audit_log_action_check
        check (action in (
            'created',      -- фиксация создана
            'transferred',  -- сменился ответственный
            'extended',     -- продлён срок
            'expired',      -- срок вышел
            'removed',      -- снята вручную
            'converted'     -- дошла до сделки
        ))
);

comment on table integration.audit_log is
    'Append-only. UPDATE и DELETE отозваны у go_user ниже в этой же миграции.';

-- «покажи историю этой фиксации»
create index if not exists audit_log_entity_idx
    on integration.audit_log (entity_type, entity_id, created_at desc);

-- «что делало это агентство за период» — по этому индексу ловят перебор
create index if not exists audit_log_actor_idx
    on integration.audit_log (actor_agency_id, created_at desc);

-- Права.
-- В bootstrap-миграции стоит `alter default privileges ... grant all`,
-- поэтому на новую таблицу go_user получил все права автоматически.
-- Здесь мы забираем лишнее обратно.
revoke update, delete on integration.audit_log from go_user;
grant  insert, select on integration.audit_log to go_user;

-- bigserial создаёт последовательность отдельным объектом,
-- и права на неё выдаются отдельно
grant usage, select on sequence integration.audit_log_id_seq to go_user;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop table if exists integration.audit_log;

-- +goose StatementEnd
