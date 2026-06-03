$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
Set-Location $repoRoot

$env:PATH += ";$env:USERPROFILE\go\bin"

protoc `
  -I "shared/pkg/grpc/proto" `
  --go_out=. `
  --go_opt=module=Broker_backend `
  --go-grpc_out=. `
  --go-grpc_opt=module=Broker_backend `
  "shared/pkg/grpc/proto/auth/v1/authv1.proto"

Write-Host "proto generated successfully"