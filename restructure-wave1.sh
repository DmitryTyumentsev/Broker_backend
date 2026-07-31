#!/usr/bin/env bash
set -euo pipefail
#
# Волна 1: границы. Только перемещения и удаления, без переименований.
# Запускать из корня репозитория на ЧИСТОМ рабочем дереве (git status пустой).
#
# После прогона `go build ./...` будет красным — это ожидаемо.
# Импорты правит GoLand (Refactor -> Move) либо руками:
#   grep -rl 'старый/путь' --include='*.go' . | xargs sed -i 's|старый/путь|новый/путь|g'
#
# Что дальше — раздел 2.4 и раздел 8 документа.

echo "==> 1. cmd наружу из internal"
for S in services/integration/brokerservice services/app/authservice; do
  mkdir -p "$S/cmd"
  git mv "$S/internal/cmd/"* "$S/cmd/"
  rmdir  "$S/internal/cmd"
done
mkdir -p services/partnerapi/cmd
git mv services/partnerapi/internal/cmd/partnerapi services/partnerapi/cmd/partnerapi
rmdir  services/partnerapi/internal/cmd

echo "==> 2. плоская раскладка сервиса фиксации"
B=services/integration/brokerservice/internal
mkdir -p "$B/repository/postgres" "$B/transport/grpc" "$B/usecase"
git mv "$B/infra/repositories/postgres/pool.go"    "$B/repository/postgres/pool.go"
git mv "$B/infra/repositories/postgres/querier.go" "$B/repository/postgres/tx.go"
git mv "$B/infra/repositories/postgres/errors.go"  "$B/repository/postgres/errors.go"
git mv "$B/infra/repositories/postgres/postgres.go" "$B/repository/postgres/postgres.go"
git mv "$B/infra/repositories/postgres/features/fixations/fixation-customer.go" \
       "$B/repository/postgres/fixation.go"
git mv "$B/transport/brokerhandler/broker-handler.go" "$B/transport/grpc/handler.go"
git mv "$B/transport/brokerhandler/errors.go"         "$B/transport/grpc/errors.go"
git mv "$B/transport/brokerhandler/errors_test.go"    "$B/transport/grpc/errors_test.go" 2>/dev/null || true
git mv "$B/usecases/"* "$B/usecase/"

echo "==> 3. удалить то, чему в сервисе фиксации не место"
# копипаста из authservice: этот сервис не хеширует пароли и не валидирует email
git rm -r -q "$B/infra/security" "$B/pkg"
# пустые каталоги после переезда
git rm -r -q "$B/infra/repositories" "$B/transport/brokerhandler" "$B/usecases" 2>/dev/null || true
# второй транспорт до своего же сервиса — раздел 9.3
git rm -q services/partnerapi/internal/clients/brokerclient/http-client.go

echo "==> 4. миграции в общий каталог по схемам"
mkdir -p migrations/bootstrap migrations/app migrations/integration
git mv services/integration/brokerservice/migrations/*.sql migrations/integration/
git mv services/app/authservice/migrations/*.sql           migrations/app/
rmdir services/integration/brokerservice/migrations services/app/authservice/migrations 2>/dev/null || true

echo
echo "Готово. Дальше руками:"
echo "  - поправить пути в makefile (BROKER_SERVICE / BROKER_MIGRATOR)"
echo "  - убрать импорт apigateway/internal/.../brokerdto из usecase и transport"
echo "  - завести usecase/command.go, transport/grpc/mapper.go"
echo "  - встроить UnimplementedBrokerServiceServer в хендлер"
echo "  - собрать fixationclient по разделу 9.4"
echo
git status --short | head -40
