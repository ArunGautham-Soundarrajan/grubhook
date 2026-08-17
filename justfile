set dotenv-load

default:
    just --list

dev:
    go run cmd/server/main.go

auth:
    go run cmd/authsetup/main.go

sqlc:
    sqlc generate

migrate:
    goose up 
