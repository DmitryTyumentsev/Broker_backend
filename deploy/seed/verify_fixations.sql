-- Постусловия сценарного сидера. Здесь намеренно нет вывода количества
-- строк: цель проверяет состав данных, не подсказывая ответ на расследование.

\set ON_ERROR_STOP on

select set_config('broker.seed_hash_key', :'key', false) as seed_hash_key \gset

do $verify$
declare
    hash_key text := current_setting('broker.seed_hash_key');
begin
    if not exists (
        select 1
          from integration.fixations
         where phone_hash = encode(hmac('79997654321', hash_key, 'sha256'), 'base64')
    ) then
        raise exception 'seed contract: current phone HMAC is missing';
    end if;

    if exists (
        select 1
          from integration.fixations
         where phone_hash = encode(hmac('79997654321', '"' || hash_key || '"', 'sha256'), 'base64')
    ) then
        raise exception 'seed contract: phone HMAC still uses a JSON-quoted key';
    end if;

    if not exists (
        select 1
          from integration.fixations f
          left join app.projects p on p.id = f.project_id
         where p.id is null
    ) then
        raise exception 'seed contract: orphaned project scenario is missing';
    end if;

    if not exists (
        select 1
          from integration.fixations
         where fix_by <> fix_for
    ) then
        raise exception 'seed contract: delegated fixation scenario is missing';
    end if;

    if not exists (
        select 1
          from integration.fixations
         group by fixed_at
        having count(*) > 25
    ) then
        raise exception 'seed contract: pagination tie scenario is missing';
    end if;

    if not exists (
        select 1
          from integration.fixations
         group by phone_hash
        having count(distinct project_id) > 1
    ) then
        raise exception 'seed contract: same client in different projects is missing';
    end if;

    -- Спорные клиенты. Проверяется наличие сценария, а не его объём:
    -- пересчёт спорных пар — это работа расследования, и подсказывать
    -- ответ постусловием сидера нельзя.
    if not exists (
        select 1
          from integration.fixations f
          join integration.fixations o
            on  o.phone_hash = f.phone_hash
            and o.project_id = f.project_id
            and o.agency_id <> f.agency_id
         where f.status = 'active'
           and o.status = 'active'
    ) then
        raise exception 'seed contract: disputed client scenario is missing';
    end if;

    -- Претендентов больше двух: правило выбора победителя обязано работать
    -- не только на паре.
    if not exists (
        select 1
          from integration.fixations
         where status = 'active'
         group by phone_hash, project_id
        having count(distinct agency_id) > 2
    ) then
        raise exception 'seed contract: three-way dispute scenario is missing';
    end if;

    -- Спор, в котором обе стороны формально активны, а сроки у обеих вышли.
    if not exists (
        select 1
          from integration.fixations f
          join integration.fixations o
            on  o.phone_hash = f.phone_hash
            and o.project_id = f.project_id
            and o.agency_id <> f.agency_id
         where f.status = 'active'
           and o.status = 'active'
           and f.expires_at < now()
           and o.expires_at < now()
    ) then
        raise exception 'seed contract: stale disputed pair scenario is missing';
    end if;

    -- Похоже на спор, но не спор: живая одна, остальные по тому же
    -- телефону и проекту закрыты. Запрос из расследования обязан их
    -- различать, значит в базе они должны быть.
    if not exists (
        select 1
          from integration.fixations
         group by phone_hash, project_id
        having count(*) filter (where status = 'active') = 1
           and count(*) filter (where status <> 'active') > 1
    ) then
        raise exception 'seed contract: closed-history decoy scenario is missing';
    end if;

    if not exists (
        select 1
          from integration.fixations
         where phone_hash in (
             encode(digest('+7 (999) 123-45-67', 'sha256'), 'base64'),
             encode(digest('89991234567', 'sha256'), 'base64')
         )
         group by project_id
        having count(distinct phone_hash) > 1
    ) then
        raise exception 'seed contract: legacy phone history is missing';
    end if;
end
$verify$;

reset broker.seed_hash_key;
\echo 'fixation seed contract: ok'
