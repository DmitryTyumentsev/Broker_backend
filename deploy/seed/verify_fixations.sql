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
