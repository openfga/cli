FROM cgr.dev/chainguard/go:latest@sha256:3348f7dc08cebd3c933c92a2cbc734c31bab5993f4b8e8d56577ba58e85cd66e AS builder

WORKDIR /app

COPY . .
RUN CGO_ENABLED=0 go build -o fga ./cmd/fga/main.go

FROM cgr.dev/chainguard/static@sha256:ee47224a2afc674c1f1089b9dea97d5ee400cf2fff3797398778450a4cfb2a8d

COPY --from=builder /app/fga /fga
ENTRYPOINT ["/fga"]
