FROM cgr.dev/chainguard/go:1.25 AS builder

WORKDIR /app

COPY . .
RUN CGO_ENABLED=0 go build -o fga ./cmd/fga/main.go

FROM cgr.dev/chainguard/static@sha256:ee47224a2afc674c1f1089b9dea97d5ee400cf2fff3797398778450a4cfb2a8d

COPY --from=builder /app/fga /fga
ENTRYPOINT ["/fga"]
