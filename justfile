set dotenv-load

default:
    just --list

dev:
    go run main.go -rod=show

sqlc:
    sqlc generate

migrate:
    goose up 
