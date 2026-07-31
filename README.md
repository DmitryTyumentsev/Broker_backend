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
