#!/usr/bin/env bash
#
# fixup.sh — добивка после restructure.sh.
#
# Чинит ровно то, что не доехало: шаг с python3 (его не оказалось в системе)
# и несколько мест, до которых первый скрипт не дотянулся.
#
# Без python. Только sed и awk. Идемпотентен: повторный запуск ничего не меняет.
#
#   cd <корень репозитория>
#   sed -i 's/\r$//' fixup.sh
#   bash fixup.sh
#
set -uo pipefail

say()  { printf '  %s\n' "$*"; }
step() { printf '\n==> %s\n' "$*"; }
skip() { printf '  = %s\n' "$*"; }
die()  { printf '\nОШИБКА: %s\n' "$*" >&2; exit 1; }

[ -f go.mod ]   || die "нет go.mod — запускай из корня репозитория"
[ -d services ] || die "нет каталога services — запускай из корня репозитория"
sed --version >/dev/null 2>&1 || die "нужен GNU sed"

BACKUP="../$(basename "$PWD")_backup_fixup_$(date +%Y%m%d_%H%M%S).tgz"
tar --exclude='./.git' --exclude='*.tgz' -czf "$BACKUP" . 2>/dev/null && say "бэкап: $BACKUP"

# ──────────────────────────────────────────────────────────────────────
# 1. ГЛАВНОЕ: загрузчик конфигов не видит контуры app/ и integration/
# ──────────────────────────────────────────────────────────────────────
# Сервисы переехали в services/app/* и services/integration/*, а viper
# ищет только services/<имя>/configs. Сейчас ни один сервис не стартует.
step "1. Пути поиска конфигов"

VL=shared/pkg/config/viper-loader.go
if [ ! -f "$VL" ]; then
  say "нет $VL — пропускаю"
elif grep -q '"app", serviceName' "$VL"; then
  skip "уже поправлен"
else
  awk '
    function cr(s) { return (substr(s, length(s)) == "\r") ? "\r" : "" }
    {
      print
      c = cr($0)
      if (!p1 && index($0, "AddConfigPath(filepath.Join(\"services\", serviceName, \"configs\"))")) {
        p1 = 1
        print "\tv.AddConfigPath(filepath.Join(\"services\", \"app\", serviceName, \"configs\"))" c
        print "\tv.AddConfigPath(filepath.Join(\"services\", \"integration\", serviceName, \"configs\"))" c
      }
      if (!p2 && index($0, "godotenv.Load(filepath.Join(\"services\", serviceName, \"configs\", env+\".env\"))")) {
        p2 = 1
        print "\t_ = godotenv.Load(filepath.Join(\"services\", \"app\", serviceName, \"configs\", env+\".env\"))" c
        print "\t_ = godotenv.Load(filepath.Join(\"services\", \"integration\", serviceName, \"configs\", env+\".env\"))" c
      }
    }
  ' "$VL" > "$VL.__tmp" && mv "$VL.__tmp" "$VL"
  grep -q '"app", serviceName' "$VL" && say "добавлены services/app/* и services/integration/*" \
    || say "! вставка не сработала, глянь $VL руками"
fi

# ──────────────────────────────────────────────────────────────────────
# 2. Префиксы переменных окружения
# ──────────────────────────────────────────────────────────────────────
# Префикс viper берёт из serviceName в верхнем регистре. serviceName
# сменился, а .env остались со старым префиксом — переменные не читаются.
step "2. Префиксы в .env"

changed=0
for f in $(grep -rl 'BROKERSERVICE_\|APIGATEWAY_' --include='*.env' --include='*.yaml' --include='*.yml' . 2>/dev/null); do
  sed -i -e 's/BROKERSERVICE_/FIXATIONSERVICE_/g' -e 's/APIGATEWAY_/PARTNERAPI_/g' "$f"
  say "$f"; changed=1
done
[ $changed -eq 0 ] && skip "уже поправлены"

# порт authservice в local.env перебивал yaml и совпадал с портом
# сервиса фиксации — два сервиса на 50052
if grep -q 'AUTHSERVICE_SERVER_PORT=50052' services/app/authservice/configs/local.env 2>/dev/null; then
  sed -i 's/AUTHSERVICE_SERVER_PORT=50052/AUTHSERVICE_SERVER_PORT=50051/' \
    services/app/authservice/configs/local.env
  say "authservice: порт 50052 -> 50051 (был занят сервисом фиксации)"
else
  skip "порт authservice уже 50051"
fi

# ──────────────────────────────────────────────────────────────────────
# 3. Пакет sharedauth: имя не совпадает с каталогом, алиас совпадает с
#    именем собственного пакета — читается тяжело
# ──────────────────────────────────────────────────────────────────────
step "3. Имена пакетов: auth и authz"

# Каталог shared/pkg/grpc/auth объявлял package sharedauth, а пакет
# shared/pkg/authz импортировался ПОД ТЕМ ЖЕ именем sharedauth.
# Имя пакета должно совпадать с каталогом, а алиас — не врать про пакет.
if grep -rq 'sharedauth' --include='*.go' . 2>/dev/null; then
  grep -rl 'sharedauth' --include='*.go' . | while read -r f; do
    sed -i \
      -e 's|sharedauth "Broker_backend/shared/pkg/authz"|"Broker_backend/shared/pkg/authz"|' \
      -e 's/\bsharedauth\./authz./g' \
      -e 's/^package sharedauth/package auth/' "$f"
    say "$f"
  done
else
  skip "уже приведены"
fi

# ──────────────────────────────────────────────────────────────────────
# 4. Линтер: каталог сгенерированного кода переехал
# ──────────────────────────────────────────────────────────────────────
step "4. .golangci.yml"
if grep -q 'shared/pkg/grpc/gen' .golangci.yml 2>/dev/null; then
  sed -i 's|- shared/pkg/grpc/gen|- gen|' .golangci.yml
  say "exclude-dirs: gen"
else
  skip "уже поправлен"
fi

# ──────────────────────────────────────────────────────────────────────
# 5. .gitignore: в нём стоит go.mod — файл, без которого проект не собрать
# ──────────────────────────────────────────────────────────────────────
step "5. .gitignore"
if [ -f .gitignore ] && grep -q '^go\.mod[[:space:]]*$' .gitignore; then
  cp .gitignore .gitignore.bak
  sed -i \
    -e '/^go\.mod[[:space:]]*$/d' \
    -e '/^services\/auth-service\//d' \
    -e '/^shared\/configs\//d' \
    .gitignore
  grep -q '^coverage\.out[[:space:]]*$' .gitignore || printf 'coverage.out\n/bin/\n' >> .gitignore
  say "убран go.mod из игнора (без него репозиторий не собирается после клона)"
  say "убраны мёртвые пути auth-service и shared/configs; старый файл в .gitignore.bak"
  say "ПРОВЕРЬ: git ls-files go.mod  — если пусто, сделай git add go.mod go.sum"
else
  skip "уже поправлен"
fi

# ──────────────────────────────────────────────────────────────────────
# 6. Мелочи
# ──────────────────────────────────────────────────────────────────────
step "6. Мелочи"
mkdir -p docs
if [ -f restructure.sh ]; then mv restructure.sh docs/restructure.sh; say "restructure.sh -> docs/"; fi
if [ -f fixup.sh ] && [ "$(dirname "$(readlink -f fixup.sh)")" = "$PWD" ]; then :; fi
[ -d _trash ] && say "напоминание: каталог _trash ещё на месте, проверь и удали"

# ──────────────────────────────────────────────────────────────────────
step "Проверка"
bad=0
grep -q '"app", serviceName' shared/pkg/config/viper-loader.go 2>/dev/null \
  && say "конфиги: пути контуров на месте" || { say "! конфиги: пути не добавились"; bad=1; }
grep -rq 'BROKERSERVICE_\|APIGATEWAY_' --include='*.env' . 2>/dev/null \
  && { say "! остались старые префиксы"; bad=1; } || say "префиксы .env: чисто"
grep -rq 'shared/pkg/grpc/gen' --include='*.yml' --include='*.yaml' . 2>/dev/null \
  && { say "! остались ссылки на старый каталог gen"; bad=1; } || say "пути gen: чисто"
grep -rq 'sharedauth' --include='*.go' . 2>/dev/null \
  && { say "! остался алиас sharedauth"; bad=1; } || say "имена пакетов: чисто"
[ $bad -eq 0 ] && say "всё сошлось"

cat <<'ENDEOF'

────────────────────────────────────────────────────────────────────────
Дальше:

  gofmt -w .
  go mod tidy
  go build ./...

Красными останутся два файла — это фича, а не скелет:
  fixationservice/internal/usecase/fixation-customer.go
  fixationservice/internal/transport/grpc/handler.go
Оба тянут DTO гейтвея из чужого internal. Лечится разведением моделей:
command.go в usecase, mapper.go в transport/grpc, встроенный
UnimplementedBrokerServiceServer в хендлере.

Проверить, что конфиги читаются:
  go run ./services/integration/fixationservice/cmd/fixationservice
падение должно быть на подключении к базе, а не на "read config file".
────────────────────────────────────────────────────────────────────────
ENDEOF
