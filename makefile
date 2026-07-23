# Makefile — положить в корень репозитория
#
# Версии плагинов зафиксированы намеренно: сгенерированный код зависит
# от версии генератора. С @latest у тебя, у коллеги и в CI получится
# разный результат, и в git полетят диффы, которых никто не делал.
#
# PROTOC_GEN_GO_VERSION должна совпадать с версией google.golang.org/protobuf
# в go.mod — тогда плагин и рантайм-библиотека гарантированно совместимы.
# Посмотреть текущую:  go list -m google.golang.org/protobuf

PROTOC_GEN_GO_VERSION      := v1.36.6
PROTOC_GEN_GO_GRPC_VERSION := v1.5.1

MIGRATOR := ./services/brokerservice/internal/cmd/migrations
SERVICE  := ./services/brokerservice/internal/cmd/brokerservice

.PHONY: tools
tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)

.PHONY: generate
generate: tools
	buf generate

.PHONY: lint
lint:
	buf lint
	golangci-lint run

.PHONY: breaking
breaking:
	buf breaking --against '.git#branch=master'

.PHONY: migrate
migrate:
	go run $(MIGRATOR)

.PHONY: run
run:
	go run $(SERVICE)

.PHONY: test
test:
	go test -race -short ./...

.PHONY: test-integration
test-integration:
	go test -race ./...

.PHONY: build
build:
	go build ./...
