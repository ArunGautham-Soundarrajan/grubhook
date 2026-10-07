FROM golang:1.26-trixie AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /out/grubhook . \
 && CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@v3.28.0 \
 && cp "$(go env GOPATH)/bin/goose" /out/goose

FROM debian:trixie-slim

# deliveroo.New launches /usr/bin/chromium via launcher.LookPath; rod adds
# --no-sandbox inside containers.
RUN apt-get update \
 && apt-get upgrade -y \
 && apt-get install -y --no-install-recommends chromium ca-certificates fonts-liberation \
 && rm -rf /var/lib/apt/lists/*

RUN useradd --create-home --uid 10001 grubhook

WORKDIR /app
RUN chown grubhook /app

COPY --from=build /out/grubhook /out/goose /usr/local/bin/
COPY db/migrations ./db/migrations

ENV GOOSE_MIGRATION_DIR=/app/db/migrations

USER grubhook

# Apply any pending migrations, then sync orders.
CMD ["sh", "-c", "goose up && exec grubhook"]
