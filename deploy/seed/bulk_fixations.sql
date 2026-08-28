-- deploy/seed/bulk_fixations.sql
--
-- Наливает в integration.fixations объём. Не демо-данные — именно объём:
-- на трёхстах строках любой запрос укладывается в единицы миллисекунд,
-- планировщик берёт seq scan и оказывается прав, и «быстро» на стенде
-- ничего не значит.
--
-- Запуск:  make seed-bulk            (по умолчанию 3 000 000 строк)
--          make seed-bulk ROWS=500000
--
-- Убрать обратно: make seed — он чистит фиксации целиком и наливает
-- демо-данные заново.
--
-- Повторный запуск подряд упрётся в уникальный индекс: номера телефонов
-- считаются от номера строки и во втором прогоне повторятся. Нужно больше
-- объёма — сначала make seed, потом make seed-bulk ROWS=...
--
-- ВАЖНО: сначала make seed, потом make seed-bulk. Скрипт раскладывает
-- строки по агентствам, сотрудникам и проектам, которые уже есть в базе;
-- на пустой базе ему нечего раскладывать.
--
-- Параметры приходят снаружи:
--   :rows  сколько строк добавить
--   :key   ключ HMAC — байты секрета без дополнительной сериализации,
--          как в fixationservice

\set ON_ERROR_STOP on
\timing on

-- Справочники разворачиваем в массивы ОДИН раз. Если оставить подзапросы
-- к app.users и app.projects внутри, они выполнятся на каждую из трёх
-- миллионов строк, и налив объёма сам станет ночным процессом.
with dict as (
    select
        (select array_agg(id        order by id) from app.users where agency_id is not null) as broker_ids,
        (select array_agg(agency_id order by id) from app.users where agency_id is not null) as broker_agencies,
        (select array_agg(id        order by id) from app.projects)                          as project_ids
)
insert into integration.fixations
    (id, fixed_at, expires_at, status, agency_id, fix_by, fix_for, project_id, phone_hash)
select
    gen_random_uuid(),
    r.fixed_at,
    r.fixed_at + interval '60 days',
    r.status,
    d.broker_agencies[r.broker],
    d.broker_ids[r.broker],
    d.broker_ids[r.broker],
    d.project_ids[r.project],
    -- Настоящий HMAC, а не случайная строка: длина, энтропия и стоимость
    -- сравнения такие же, как у боевых данных. На кривом хэше EXPLAIN
    -- врёт в свою пользу.
    --
    -- Диапазон номеров 795XXXXXXXX намеренно не пересекается с тем, что
    -- раздаёт сидер: иначе объём упирается в частичный уникальный индекс
    -- (phone_hash, project_id) и налив падает на середине.
    encode(hmac('795' || lpad(g::text, 8, '0'), :'key', 'sha256'), 'base64')
from generate_series(1, :rows) as g
cross join dict as d
cross join lateral (
    select
        1 + (g % array_length(d.broker_ids, 1))  as broker,
        1 + (g % array_length(d.project_ids, 1)) as project,
        -- Разброс на три года назад, начиная с 200 дней: демо-данные лежат
        -- в последних 180 днях и остаются самыми свежими. Иначе первую
        -- страницу списка занимает шум, и глазами проверить нечего.
        now()
            - interval '200 days'
            - ((g % 1000) * interval '1 day')
            - ((g % 86400) * interval '1 second')                                     as fixed_at,
        -- Перекос в active такой же, как у сидера.
        (array['active', 'active', 'active', 'converted', 'expired', 'removed'])[1 + (g % 6)] as status
) as r;

-- Без этого планировщик считает, что в таблице по-прежнему триста строк,
-- и EXPLAIN показывает план, которого в проде не будет.
analyze integration.fixations;
