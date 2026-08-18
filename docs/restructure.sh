#!/usr/bin/env bash
#
# restructure.sh — приведение структуры проекта к целевой архитектуре.
#
# ДЕЛАЕТ: переносит и переименовывает каталоги и файлы, чинит package-клаузы,
#         пути импортов и квалификаторы пакетов, правит makefile / buf /
#         prometheus / compose / конфиги, добавляет bootstrap-миграцию и README.
#
# НЕ ДЕЛАЕТ: не трогает бизнес-логику, не удаляет комментарии и вопросы,
#            не переименовывает типы и функции внутри фич,
#            не трогает .proto и сгенерированный код (это отдельный шаг ниже).
#
# ИДЕМПОТЕНТЕН: повторный запуск ничего не ломает и не плодит дубли.
#
# Запуск:
#   cd <корень репозитория>
#   sed -i 's/\r$//' restructure.sh      # на случай, если файл приехал с CRLF
#   bash restructure.sh
#
set -uo pipefail

say()  { printf '  %s\n' "$*"; }
step() { printf '\n==> %s\n' "$*"; }
warn() { printf '  ! %s\n' "$*"; }
die()  { printf '\nОШИБКА: %s\n' "$*" >&2; exit 1; }

# ──────────────────────────────────────────────────────────────────────
# 0. Проверки и бэкап
# ──────────────────────────────────────────────────────────────────────
[ -f go.mod ]   || die "нет go.mod — запускай из корня репозитория"
[ -d services ] || die "нет каталога services — запускай из корня репозитория"
sed --version >/dev/null 2>&1 || die "нужен GNU sed (в WSL стоит из коробки)"

MODULE="$(sed -n 's/^module[[:space:]]\+//p' go.mod | tr -d '\r' | head -1)"
[ -n "$MODULE" ] || die "не смог прочитать module из go.mod"
M="$MODULE"
say "модуль: $M"

BACKUP="../$(basename "$PWD")_backup_$(date +%Y%m%d_%H%M%S).tgz"
step "Бэкап"
tar --exclude='./.git' --exclude='./_trash' --exclude='*.tgz' -czf "$BACKUP" . 2>/dev/null \
  && say "$BACKUP" || warn "бэкап не создался, продолжаю"

mkdir -p _trash

# ──────────────────────────────────────────────────────────────────────
# Хелперы
# ──────────────────────────────────────────────────────────────────────
mvpath() {                       # переносит, если источник есть; сливает, если цель есть
  local src="$1" dst="$2"
  [ -e "$src" ] || return 0
  if [ -e "$dst" ]; then
    if [ -d "$src" ] && [ -d "$dst" ]; then
      shopt -s dotglob nullglob
      local items=("$src"/*)
      [ ${#items[@]} -gt 0 ] && mv -n "${items[@]}" "$dst"/ 2>/dev/null
      shopt -u dotglob nullglob
      rmdir "$src" 2>/dev/null && say "слил $src -> $dst"
    fi
    return 0
  fi
  mkdir -p "$(dirname "$dst")"
  mv "$src" "$dst" && say "$src -> $dst"
}

trash() {
  local p="$1"; [ -e "$p" ] || return 0
  mkdir -p "_trash/$(dirname "$p")"
  mv "$p" "_trash/$p" 2>/dev/null && say "в _trash: $p"
}

esc() { printf '%s' "$1" | sed 's/[][\.*^$/|&]/\\&/g'; }

repl() {                         # repl <старое> <новое> [расширение...]
  local old="$1" new="$2"; shift 2
  local exts=("$@"); [ ${#exts[@]} -eq 0 ] && exts=("*.go")
  local inc=()
  local e; for e in "${exts[@]}"; do inc+=(--include="$e"); done
  grep -rlF "${inc[@]}" --exclude-dir=_trash --exclude-dir=.git -- "$old" . 2>/dev/null \
    | xargs -r sed -i "s|$(esc "$old")|$(esc "$new")|g"
}

pkg() {                          # pkg <каталог> <имя пакета>
  local dir="$1" name="$2"; [ -d "$dir" ] || return 0
  local f changed=0
  for f in "$dir"/*.go; do
    [ -e "$f" ] || continue
    sed -i "s/^package [A-Za-z0-9_]\+/package $name/" "$f" && changed=1
  done
  [ $changed -eq 1 ] && say "package $name  <- $dir"
  return 0
}

dedupe_imports() {               # убирает точные дубли строк внутри import ( ... )
  local f="$1"; [ -f "$f" ] || return 0
  awk '
    BEGIN { inimp = 0 }
    /^import \(/ { inimp = 1; print; next }
    inimp && /^\)/ { inimp = 0; delete seen; print; next }
    {
      if (inimp) {
        k = $0; sub(/\r$/, "", k); gsub(/^[ \t]+/, "", k)
        if (k != "" && (k in seen)) next
        if (k != "") seen[k] = 1
      }
      print
    }
  ' "$f" > "$f.__tmp" && mv "$f.__tmp" "$f"
}

FIX=services/integration/fixationservice
PAPI=services/integration/partnerapi
AUTH=services/app/authservice

# ──────────────────────────────────────────────────────────────────────
# 1. Каталоги сервисов и точки входа
# ──────────────────────────────────────────────────────────────────────
step "1. Каталоги сервисов"
mvpath services/integration/brokerservice "$FIX"
mvpath services/apigateway               "$PAPI"
mvpath services/brokerservice            "$FIX"
mvpath "$FIX/cmd/brokerservice" "$FIX/cmd/fixationservice"
mvpath "$FIX/cmd/migrations"    "$FIX/cmd/migrator"
mvpath "$PAPI/cmd/apigateway"   "$PAPI/cmd/partnerapi"

# ──────────────────────────────────────────────────────────────────────
# 2. Внутренняя раскладка: один шаблон на все сервисы
# ──────────────────────────────────────────────────────────────────────
step "2. Внутренняя раскладка"

mvpath "$FIX/internal/infra/repositories/postgres/features/fixations/fixation-customer.go" \
       "$FIX/internal/repository/postgres/fixation.go"
mvpath "$FIX/internal/infra/repositories/postgres" "$FIX/internal/repository/postgres"
mvpath "$FIX/internal/infra/cache/redis"           "$FIX/internal/infra/redis"
mvpath "$FIX/internal/transport/brokerhandler"     "$FIX/internal/transport/grpc"
mvpath "$FIX/internal/usecases"                    "$FIX/internal/usecase"

mvpath "$AUTH/internal/infra/repositories/postgres" "$AUTH/internal/repository/postgres"
mvpath "$AUTH/internal/infra/cache/redis"           "$AUTH/internal/infra/redis"
mvpath "$AUTH/internal/transport/grpcauthhandler"   "$AUTH/internal/transport/grpc"
mvpath "$AUTH/internal/usecases"                    "$AUTH/internal/usecase"
mvpath "$AUTH/internal/pkg/validators"              "$AUTH/internal/validators"

mvpath "$PAPI/internal/infra/cache/redis"                      "$PAPI/internal/infra/redis"
mvpath "$PAPI/internal/clients/brokerclient"                   "$PAPI/internal/clients/fixationclient"
mvpath "$PAPI/internal/transport/http/dto/brokerdto"           "$PAPI/internal/transport/http/dto/fixationdto"
mvpath "$PAPI/internal/transport/http/handlers/brokerhandlers" "$PAPI/internal/transport/http/handlers/fixationhandlers"
mvpath "$PAPI/internal/transport/http/broker-routes.go"        "$PAPI/internal/transport/http/fixation-routes.go"

step "2.1 Имена файлов"
mvpath "$FIX/internal/transport/grpc/broker-handler.go"  "$FIX/internal/transport/grpc/handler.go"
mvpath "$AUTH/internal/transport/grpc/grpchandler.go"    "$AUTH/internal/transport/grpc/handler.go"
mvpath "$AUTH/internal/usecase/cmd.go"                   "$AUTH/internal/usecase/commands.go"
mvpath "$PAPI/internal/clients/fixationclient/grpc-client.go" "$PAPI/internal/clients/fixationclient/client.go"
mvpath "$PAPI/internal/clients/authclient/auth-client.go"     "$PAPI/internal/clients/authclient/client.go"
mvpath "$PAPI/internal/transport/http/handlers/fixationhandlers/broker-handlers.go" \
       "$PAPI/internal/transport/http/handlers/fixationhandlers/handler.go"
mvpath "$PAPI/internal/transport/http/handlers/authhandlers/auth-handlers.go" \
       "$PAPI/internal/transport/http/handlers/authhandlers/handler.go"
mvpath "$PAPI/internal/transport/http/dto/authdto/auth-dto.go" \
       "$PAPI/internal/transport/http/dto/authdto/auth.go"

# ──────────────────────────────────────────────────────────────────────
# 3. Общее — в shared, контракты — в корень
# ──────────────────────────────────────────────────────────────────────
step "3. shared и контракты"
if [ ! -d shared/pkg/clock ]; then
  if   [ -d "$FIX/internal/infra/clock" ];  then mvpath "$FIX/internal/infra/clock"  shared/pkg/clock
  elif [ -d "$AUTH/internal/infra/clock" ]; then mvpath "$AUTH/internal/infra/clock" shared/pkg/clock
  fi
fi
rm -rf "$FIX/internal/infra/clock" "$AUTH/internal/infra/clock" "$PAPI/internal/infra/clock" 2>/dev/null
mvpath shared/pkg/grpc/proto proto
mvpath shared/pkg/grpc/gen   gen

# ──────────────────────────────────────────────────────────────────────
# 4. package-клаузы
# ──────────────────────────────────────────────────────────────────────
step "4. package-клаузы"
pkg "$FIX/internal/repository/postgres"  postgres
pkg "$FIX/internal/transport/grpc"       grpc
pkg "$FIX/internal/usecase"              usecase
pkg "$AUTH/internal/repository/postgres" postgres
pkg "$AUTH/internal/repository/postgres/users"    users
pkg "$AUTH/internal/repository/postgres/sessions" sessions
pkg "$AUTH/internal/transport/grpc"      grpc
pkg "$AUTH/internal/usecase"             usecase
pkg "$PAPI/internal/clients/fixationclient"                   fixationclient
pkg "$PAPI/internal/transport/http/dto/fixationdto"           fixationdto
pkg "$PAPI/internal/transport/http/handlers/fixationhandlers" fixationhandlers

# ──────────────────────────────────────────────────────────────────────
# 5. Пути импортов (длинные раньше коротких)
# ──────────────────────────────────────────────────────────────────────
step "5. Пути импортов"
repl "$M/services/integration/brokerservice" "$M/services/integration/fixationservice"
repl "$M/services/brokerservice"             "$M/services/integration/fixationservice"
repl "$M/services/apigateway"                "$M/services/integration/partnerapi"
repl "$M/services/integration/fixationservice/internal/infra/clock" "$M/shared/pkg/clock"
repl "$M/services/app/authservice/internal/infra/clock"             "$M/shared/pkg/clock"
repl "/internal/infra/repositories/postgres/features/fixations" "/internal/repository/postgres"
repl "/internal/infra/repositories/postgres" "/internal/repository/postgres"
repl "/internal/infra/cache/redis"           "/internal/infra/redis"
repl "/internal/transport/brokerhandler"     "/internal/transport/grpc"
repl "/internal/transport/grpcauthhandler"   "/internal/transport/grpc"
repl "/internal/usecases"                    "/internal/usecase"
repl "/internal/pkg/validators"              "/internal/validators"
repl "/internal/clients/brokerclient"        "/internal/clients/fixationclient"
repl "/dto/brokerdto"                        "/dto/fixationdto"
repl "/handlers/brokerhandlers"              "/handlers/fixationhandlers"
repl "$M/shared/pkg/grpc/gen"                "$M/gen"
repl "$M/shared/pkg/auth\""                  "$M/shared/pkg/authz\""   # был битый путь
say "готово"

# ──────────────────────────────────────────────────────────────────────
# 6. Квалификаторы пакетов
# ──────────────────────────────────────────────────────────────────────
step "6. Квалификаторы пакетов"
repl "usecases."       "usecase."
repl "brokerclient."   "fixationclient."
repl "brokerdto."      "fixationdto."
repl "brokerhandlers." "fixationhandlers."

# fixation.go переехал в сам пакет postgres: свой пакет не квалифицируют,
# поэтому здесь алиас и импорт убираются целиком, а не переименовываются
for f in "$FIX"/internal/repository/postgres/*.go; do
  [ -e "$f" ] || continue
  grep -q 'postgres2' "$f" || continue
  sed -i -e '/postgres2 "/d' -e 's/postgres2\.//g' "$f"
  say "убрана самоквалификация: $f"
done

# во всех остальных файлах алиас превращается в имя пакета
repl 'postgres2 "'     '"'
repl "postgres2."      "postgres."
repl "fixations.NewRepository" "postgres.NewRepository"
repl '"postgres.features.fixations.' '"postgres.fixation.'
say "готово"

# ──────────────────────────────────────────────────────────────────────
# 7. Composition root: алиас транспорта и дубли импортов
# ──────────────────────────────────────────────────────────────────────
step "7. Composition root"
alias_transport() {              # transport/grpc конфликтует с google.golang.org/grpc
  local f="$1" oldq="$2"; [ -f "$f" ] || return 0
  grep -q 'grpctransport ' "$f" || \
    sed -i "s|^\(\s*\)\"$(esc "$M")/\(services/[a-z/]*\)/internal/transport/grpc\"|\1grpctransport \"$(esc "$M")/\2/internal/transport/grpc\"|" "$f"
  sed -i "s/\b${oldq}\./grpctransport./g" "$f"
  say "алиас grpctransport: $f"
}
alias_transport "$FIX/internal/app/app.go"  brokerhandler
alias_transport "$AUTH/internal/app/app.go" grpcauthhandler

# самоимпорт после слияния пакета fixations в postgres
F="$FIX/internal/repository/postgres/fixation.go"
[ -f "$F" ] && sed -i "\|$(esc "$M")/services/integration/fixationservice/internal/repository/postgres\"|d" "$F"

# точные дубли импортов после схлопывания путей
while IFS= read -r f; do dedupe_imports "$f"; done < <(find services shared -name '*.go' 2>/dev/null)
say "дубли импортов убраны"

# ──────────────────────────────────────────────────────────────────────
# 8. Конфиги: пути поиска, пути миграций, имена
# ──────────────────────────────────────────────────────────────────────
step "8. Конфиги"

VL=shared/pkg/config/viper-loader.go
if [ -f "$VL" ] && ! grep -q '"app", serviceName' "$VL"; then
  if command -v python3 >/dev/null 2>&1; then
    python3 - "$VL" <<'PYEOF'
import sys, io
p = sys.argv[1]
s = io.open(p, encoding='utf-8', newline='').read()
nl = '\r\n' if '\r\n' in s else '\n'
a = 'v.AddConfigPath(filepath.Join("services", serviceName, "configs"))'
if a in s:
    s = s.replace(a, a + nl +
        '\tv.AddConfigPath(filepath.Join("services", "app", serviceName, "configs"))' + nl +
        '\tv.AddConfigPath(filepath.Join("services", "integration", serviceName, "configs"))', 1)
b = 'godotenv.Load(filepath.Join("services", serviceName, "configs", env+".env"))'
if b in s:
    s = s.replace(b, b + nl +
        '\t_ = godotenv.Load(filepath.Join("services", "app", serviceName, "configs", env+".env"))' + nl +
        '\t_ = godotenv.Load(filepath.Join("services", "integration", serviceName, "configs", env+".env"))', 1)
io.open(p, 'w', encoding='utf-8', newline='').write(s)
print("  viper-loader: добавлены пути services/app/* и services/integration/*")
PYEOF
  else
    warn "нет python3 — добавь в $VL два AddConfigPath вручную (см. чеклист)"
  fi
else
  say "viper-loader уже поправлен"
fi

# миграции переехали в общий каталог
repl 'services/brokerservice/migrations' 'migrations/integration' '*.go' '*.yaml' '*.yml' '*.env'
repl 'services/authservice/migrations'   'migrations/app'         '*.go' '*.yaml' '*.yml' '*.env'
repl '"services", "brokerservice", "migrations"' '"migrations", "integration"' '*.go'
repl '"services", "authservice", "migrations"'   '"migrations", "app"'         '*.go'

# namespace метрик гейтвея — до общего переименования
repl 'Namespace: "brokerservice"' 'Namespace: "partnerapi"' '*.go'

# ключ конфига до сервиса фиксации
repl 'broker_grpc'      'fixation_grpc'      '*.go' '*.yaml' '*.yml' '*.env'
repl 'BrokerGRPCConfig' 'FixationGRPCConfig' '*.go'
repl 'BrokerGRPC'       'FixationGRPC'       '*.go'

# общее переименование словом: логи, service_name, адреса, echo в makefile
repl 'brokerservice' 'fixationservice' '*.go' '*.yaml' '*.yml' '*.env' 'makefile' 'Makefile'
repl 'apigateway'    'partnerapi'      '*.go' '*.yaml' '*.yml' '*.env' 'makefile' 'Makefile'
say "имена приведены"

# ──────────────────────────────────────────────────────────────────────
# 9. Инструменты сборки и инфраструктура
# ──────────────────────────────────────────────────────────────────────
step "9. makefile, buf, prometheus, compose"
sed -i \
  -e 's|^BROKER_SERVICE .*|FIXATION_SERVICE  := ./services/integration/fixationservice/cmd/fixationservice|' \
  -e 's|^BROKER_MIGRATOR .*|FIXATION_MIGRATOR := ./services/integration/fixationservice/cmd/migrator|' \
  -e 's|\$(BROKER_SERVICE)|$(FIXATION_SERVICE)|g' \
  -e 's|\$(BROKER_MIGRATOR)|$(FIXATION_MIGRATOR)|g' \
  makefile 2>/dev/null && say "makefile"

sed -i 's|path: shared/pkg/grpc/proto|path: proto|' buf.yaml 2>/dev/null && say "buf.yaml"
sed -i -e "s|value: $(esc "$M")/shared/pkg/grpc/gen|value: $M/gen|" \
       -e 's|out: shared/pkg/grpc/gen|out: gen|' buf.gen.yaml 2>/dev/null && say "buf.gen.yaml"

sed -i -e "s|host.docker.internal:8081|host.docker.internal:8080|" \
       deploy/prometheus.yml 2>/dev/null && say "prometheus.yml"

sed -i 's|pg_isready -U fixation -d fixation|pg_isready -U postgres -d fixation|' \
       docker-compose.yml 2>/dev/null && say "docker-compose healthcheck"

# ──────────────────────────────────────────────────────────────────────
# 10. Новые каталоги, bootstrap-миграция, README
# ──────────────────────────────────────────────────────────────────────
step "10. Каталоги под будущие фичи"
mkdir -p migrations/bootstrap build tools docs
for d in build tools docs; do [ -e "$d/.gitkeep" ] || touch "$d/.gitkeep"; done
say "build/ tools/ docs/ migrations/bootstrap/"

BOOT=migrations/bootstrap/0001_schemas_and_roles.sql
if [ ! -f "$BOOT" ]; then
cat > "$BOOT" <<'SQLEOF'
-- Схемы и роли двух контуров. Катится ОДИН РАЗ суперпользователем,
-- до всех остальных миграций:
--   psql -U postgres -d broker -f migrations/bootstrap/0001_schemas_and_roles.sql

-- +goose Up
-- +goose StatementBegin
create schema if not exists app;           -- владелец: монолит (Laravel / authservice)
create schema if not exists integration;   -- владелец: Go-контур

do $$
begin
  if not exists (select 1 from pg_roles where rolname = 'app_user') then
    create role app_user login password 'app_user';
  end if;
  if not exists (select 1 from pg_roles where rolname = 'go_user') then
    create role go_user login password 'go_user';
  end if;
end
$$;

grant usage on schema app, integration to app_user, go_user;

-- монолит: полные права на свою схему, только чтение чужой
grant all privileges on all tables in schema app         to app_user;
grant select         on all tables in schema integration to app_user;

-- Go: зеркально
grant all privileges on all tables in schema integration to go_user;
grant select         on all tables in schema app         to go_user;

-- то же для таблиц, которые появятся позже
alter default privileges in schema app         grant all    on tables to app_user;
alter default privileges in schema app         grant select on tables to go_user;
alter default privileges in schema integration grant all    on tables to go_user;
alter default privileges in schema integration grant select on tables to app_user;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop schema if exists integration cascade;
drop schema if exists app cascade;
-- +goose StatementEnd
SQLEOF
say "создан $BOOT"
fi

# ──────────────────────────────────────────────────────────────────────
# 11. Легаси — в _trash
# ──────────────────────────────────────────────────────────────────────
step "11. Легаси"
trash READ.me
trash restructure-wave1.sh
trash shared/configs
trash scripts/deploy.sh
trash scripts/test-all.sh
trash scripts/generate-protoc.sh
trash scripts/generate-protoc.ps1

grep -q '^_trash/' .gitignore 2>/dev/null || printf '\n_trash/\n*.tgz\n' >> .gitignore

if [ ! -f README.md ]; then
cat > README.md <<'MDEOF'
# ЛК брокера застройщика — Go-контур

Гибрид Laravel + Go. В репозитории только Go-часть; роль монолита играют заглушки.

## Кто за что отвечает

| Сервис | Контур | Роль |
|---|---|---|
| `services/app/authservice` | монолит | Заглушка вместо Laravel Sanctum: выпускает токены. Владеет схемой `app` |
| `services/integration/partnerapi` | Go | Публичный вход: `/internal/v1` для монолита, `/partner/v1` для CRM агентств |
| `services/integration/fixationservice` | Go | Горячий путь фиксации клиента. Владеет схемой `integration` |

Точка входа для браузера — у монолита (Laravel). Go торчит наружу только через `partnerapi`.
`fixationservice` наружу не смотрит: только gRPC внутри периметра.

## Раскладка сервиса

```
cmd/<бинарь>/main.go     запуск, 10 строк
configs/                 local.yaml, dev.yaml
internal/
  app/                   composition root
  config/
  domain/                entity + доменные ошибки
  usecase/               сценарии и порты (интерфейсы зависимостей)
  repository/postgres/   реализация портов доступа к данным
  infra/                 технические адаптеры: redis, security
  transport/grpc|http/   хендлеры и мапперы
```

## База

Одна база, две схемы. Границу держат гранты, а не договорённости:
`app_user` — полные права на `app.*`, только чтение `integration.*`; `go_user` — зеркально.

```
migrations/bootstrap    схемы и роли, один раз суперюзером
migrations/app          катит authservice/cmd/migrator
migrations/integration  катит fixationservice/cmd/migrator
```

## Запуск

```
make up          postgres, redis, jaeger, prometheus, grafana
make migrate     миграции
make run         fixationservice
make generate    proto -> Go
make lint test
```
MDEOF
say "создан README.md"
fi

# ──────────────────────────────────────────────────────────────────────
# 12. Уборка
# ──────────────────────────────────────────────────────────────────────
step "12. Уборка пустых каталогов"
find services shared -type d -empty -delete 2>/dev/null
say "готово"

cat <<'ENDEOF'

────────────────────────────────────────────────────────────────────────
СТРУКТУРА ПРИВЕДЕНА. Дальше руками — то, что скрипт трогать не должен.

1.  gofmt -w .   (и goimports, если стоит — порядок импортов поедет)
    go mod tidy

2.  Два незаконных импорта — чинишь ты, это фича, а не скелет:
      fixationservice/internal/usecase/fixation-customer.go
      fixationservice/internal/transport/grpc/handler.go
    Оба тянут DTO гейтвея из чужого internal — компилятор такое не пропустит.
      - в usecase завести command.go: юзкейс принимает свою команду
      - в transport/grpc завести mapper.go: proto -> command, entity -> proto
      - в handler встроить UnimplementedBrokerServiceServer, сигнатуры из proto

3.  Контракт broker.v1 -> fixation.v1 требует перегенерации, поэтому отдельно,
    когда `make generate` заработает:
      sed -i 's/package broker\.v1/package fixation.v1/; s/BrokerService/FixationService/g; s/broker_id/agency_id/g' proto/broker/v1/*.proto
      mv proto/broker proto/fixation && mv proto/fixation/v1/brokerv1.proto proto/fixation/v1/fixation.proto
      rm -rf gen/broker && make generate
      grep -rl 'gen/broker/v1' --include='*.go' . | xargs sed -i 's|gen/broker/v1|gen/fixation/v1|g; s|brokerv1|fixationv1|g'

4.  Имена типов внутри фич не трогал — переименуй в IDE, когда будешь в файле:
      BrokerHandler -> FixationHandler, registerBrokerRoutes -> registerFixationRoutes,
      GRPCClient -> Client, FixationCustomer -> Fixation.

5.  DSN по ролям в configs/local.yaml, и накатить migrations/bootstrap:
      fixationservice: postgres://go_user:...@localhost:5432/broker?search_path=integration,public
      authservice:     postgres://app_user:...@localhost:5432/broker?search_path=app,public

6.  Проверь _trash/ и удали. Полный бэкап дерева лежит рядом с репозиторием.
────────────────────────────────────────────────────────────────────────
ENDEOF
