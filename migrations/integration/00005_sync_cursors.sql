-- migrations/integration/00005_sync_cursors.sql
--
-- Курсоры инкрементальной синхронизации с внешними системами.
--
-- Задача: при каждом проходе воркера забирать у amoCRM / Profitbase / 1С
-- только то, что изменилось с прошлого раза. Полная выгрузка на каждом
-- цикле — это и лишний трафик, и гарантированный упор в их rate limit.
--
-- Курсор обязан лежать в базе, а не в памяти процесса: под нагрузкой
-- воркер перезапускают, и после рестарта он должен продолжить с того же
-- места, а не начать историю заново.

-- +goose Up
-- +goose StatementBegin

create table if not exists integration.sync_cursors (
    -- Какая внешняя система: 'amocrm', 'profitbase', '1c'.
    source          text        not null,

    -- Что именно синхронизируем внутри неё: 'leads', 'property', 'contacts'.
    -- Пара (source, resource) — естественный ключ: один курсор на поток.
    -- Суррогатного id нет намеренно, он тут ничего не добавляет.
    resource        text        not null,

    -- Позиция. Держим два вида, потому что API у всех разные:
    --
    -- position_at — «отдай изменённое после этого момента» (amoCRM умеет
    -- фильтр по updated_at). Время внешней системы, не наше: сравнивать
    -- его с now() нельзя, часы разъезжаются.
    position_at     timestamptz,

    -- position_token — непрозрачный курсор пагинации, если API отдаёт
    -- его вместо времени (Profitbase так и делает). Для нас это строка,
    -- которую нельзя разбирать и о структуре которой нельзя гадать.
    position_token  text,

    -- Когда проход завершился успешно. По разнице с now() строится
    -- алерт «синхронизация встала» — самый полезный сигнал по всей
    -- интеграции: расхождение с CRM начинается именно здесь.
    last_success_at timestamptz,

    -- Когда проход в последний раз упал, и с чем. Текстом, без разбора
    -- на категории: решение «повторять или нет» принимает код воркера.
    last_error_at   timestamptz,
    last_error      text,

    updated_at      timestamptz not null default now(),

    constraint sync_cursors_pkey primary key (source, resource)
);

comment on table integration.sync_cursors is
    'Позиция инкрементальной синхронизации по каждому внешнему источнику.';

-- «Где мы отстали» — один запрос на всю таблицу, сортировка по времени
-- последнего успеха. Таблица маленькая (единицы строк), поэтому больше
-- индексов здесь не нужно: любой из них Postgres всё равно проигнорирует
-- в пользу seq scan.
create index if not exists sync_cursors_last_success_idx
    on integration.sync_cursors (last_success_at);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

drop table if exists integration.sync_cursors;

-- +goose StatementEnd
