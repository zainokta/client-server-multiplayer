FROM golang:1.23.3-alpine3.20 AS build

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o service .

FROM debian:bookworm-slim

WORKDIR /app

COPY --from=build /app/service .

ENTRYPOINT ["/app/service"]