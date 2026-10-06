set dotenv-load

default:
    just --list

dev:
    go run main.go

sqlc:
    sqlc generate

migrate:
    goose up 
