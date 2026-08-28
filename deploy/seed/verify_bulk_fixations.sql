-- Постусловие объёмной фикстуры. Проверяется край диапазона, а не число
-- строк в таблице: сценарные данные уже лежат там же и входят в итог.

\set ON_ERROR_STOP on

select set_config('broker.seed_hash_key', :'key', false) as seed_hash_key \gset
select set_config('broker.seed_rows', :'rows', false) as seed_rows \gset

do $verify$
declare
    hash_key     text := current_setting('broker.seed_hash_key');
    boundary_row text := current_setting('broker.seed_rows');
    phone        text := '795' || lpad(boundary_row, 8, '0');
begin
    if not exists (
        select 1
          from integration.fixations
         where phone_hash = encode(hmac(phone, hash_key, 'sha256'), 'base64')
    ) then
        raise exception 'bulk seed contract: boundary fixture is missing';
    end if;
end
$verify$;

reset broker.seed_hash_key;
reset broker.seed_rows;
\echo 'bulk fixation seed contract: ok'
