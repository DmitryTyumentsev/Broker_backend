## Среда выполнения

Проект открыт в GoLand на Windows, но все команды сборки, тестирования,
линтеров, миграций, сидирования и Docker необходимо выполнять в Ubuntu WSL.

Используй:

`wsl.exe -d Ubuntu --cd /mnt/c/Users/dvtyu/GolandProjects/Broker_backend -- bash -lc "<команда>"`

Не запускай Windows-версию Go:
`C:\Users\dvtyu\sdk\go1.26.3\bin\go.exe`

Не запускай созданные Go файлы `*.test.exe` в Windows.

Примеры:

```text
wsl.exe -d Ubuntu --cd /mnt/c/Users/dvtyu/GolandProjects/Broker_backend -- bash -lc "go test ./..."
wsl.exe -d Ubuntu --cd /mnt/c/Users/dvtyu/GolandProjects/Broker_backend -- bash -lc "make test"
wsl.exe -d Ubuntu --cd /mnt/c/Users/dvtyu/GolandProjects/Broker_backend -- bash -lc "make migrate"
```
