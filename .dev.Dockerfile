FROM cgr.dev/chainguard/go:1.20@sha256:8454bbbb4061bd8fed6ce0b6de0d08c0a6037fe33e136b3f16dba31a68b9b3b6 AS builder

WORKDIR /app

COPY . .
RUN CGO_ENABLED=0 go build -o fga ./cmd/fga/main.go

FROM cgr.dev/chainguard/static@sha256:ee47224a2afc674c1f1089b9dea97d5ee400cf2fff3797398778450a4cfb2a8d

COPY --from=builder /app/fga /fga
ENTRYPOINT ["/fga"]
