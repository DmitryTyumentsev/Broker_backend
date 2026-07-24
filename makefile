# Makefile — каталог операций проекта.
#
# Зачем: команды перестают жить в голове и в истории терминала.
# Новый человек в команде читает Makefile и видит всё, что можно делать.
# CI вызывает те же цели — одна точка правды.
#
# ВАЖНО: отступы — ТАБУЛЯЦИЯ, не пробелы. Иначе "missing separator".

# ── Версии инструментов ────────────────────────────────────────────────
# Пиннинг намеренный: с @latest у тебя, у коллеги и в CI получится
# разный сгенерированный код, и в git полетят диффы, которых никто не писал.
#
# PROTOC_GEN_GO_VERSION должна совпадать с версией google.golang.org/protobuf
# в go.mod — сгенерированный .pb.go содержит проверку совместимости версий.
# Посмотреть текущую:  go list -m google.golang.org/protobuf

PROTOC_GEN_GO_VERSION      := v1.36.11
PROTOC_GEN_GO_GRPC_VERSION := v1.5.1

BROKER_SERVICE  := ./services/brokerservice/internal/cmd/brokerservice
BROKER_MIGRATOR := ./services/brokerservice/internal/cmd/migrations
GRPC_ADDR       := localhost:50052

# .PHONY говорит make: это команды, а не имена файлов.
# Без него make увидит файл с таким именем и решит, что делать нечего.
.PHONY: help tools generate lint fmt vet test test-integration cover \
        build run migrate up down logs psql grpc-list grpc-fix vuln breaking

# ── Справка ────────────────────────────────────────────────────────────
help:
	@echo "Инфраструктура:"
	@echo "  make up                поднять postgres/redis/jaeger/prometheus/grafana"
	@echo "  make down              остановить"
	@echo "  make psql              консоль postgres"
	@echo ""
	@echo "Разработка:"
	@echo "  make migrate           накатить миграции"
	@echo "  make run               запустить brokerservice"
	@echo "  make generate          proto -> Go"
	@echo ""
	@echo "Проверки:"
	@echo "  make lint              golangci-lint + buf lint"
	@echo "  make test              быстрые тесты (-short)"
	@echo "  make test-integration  все тесты, поднимает Postgres в Docker"
	@echo "  make vuln              скан уязвимостей"
	@echo ""
	@echo "Проверка ручки вручную:"
	@echo "  make grpc-list         какие методы есть на сервере"
	@echo "  make grpc-fix          дёрнуть NewFixationCustomer"

# ── Инструменты ────────────────────────────────────────────────────────
tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)
	go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

# "generate: tools" = перед generate выполни tools.
# Не надо помнить порядок — make сам разрулит.
generate: tools
	buf generate

# ── Инфраструктура ─────────────────────────────────────────────────────
up:
	docker compose up -d
	@echo "Jaeger      http://localhost:16686"
	@echo "Prometheus  http://localhost:9090"
	@echo "Grafana     http://localhost:3000  (admin/admin)"

down:
	docker compose down

logs:
	docker compose logs -f

psql:
	docker compose exec postgres psql -U broker -d broker

# ── Разработка ─────────────────────────────────────────────────────────
build:
	go build ./...

run:
	go run $(BROKER_SERVICE)

migrate:
	go run $(BROKER_MIGRATOR)

# ── Проверки ───────────────────────────────────────────────────────────
fmt:
	gofmt -w .

vet:
	go vet ./...

lint:
	golangci-lint run
	buf lint

# -short пропускает интеграционные (в них стоит testing.Short() -> Skip)
test:
	go test -race -short ./...

test-integration:
	go test -race ./...

cover:
	go test -short -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

vuln:
	govulncheck ./...

breaking:
	buf breaking --against '.git#branch=main'

# ── Проверка ручки руками ──────────────────────────────────────────────
# Работает благодаря reflection, включённой в app.go для локалки.
grpc-list:
	grpcurl -plaintext $(GRPC_ADDR) list
	grpcurl -plaintext $(GRPC_ADDR) list broker.v1.BrokerService

grpc-fix:
	grpcurl -plaintext -d '{ \
	  "broker_id":   "11111111-1111-1111-1111-111111111111", \
	  "customer_id": "22222222-2222-2222-2222-222222222222", \
	  "fix_for":     "33333333-3333-3333-3333-333333333333", \
	  "fixed_by":    "44444444-4444-4444-4444-444444444444" \
	}' $(GRPC_ADDR) broker.v1.BrokerService/NewFixationCustomer

# Проверить, что записалось: времена не перепутаны, статус активный
check-db:
	docker compose exec postgres psql -U broker -d broker -c \
	  "select id, status, fixed_at, expires_at, expires_at > fixed_at as period_ok \
	   from fixation_customers order by fixed_at desc limit 5;"
