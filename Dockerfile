FROM golang AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -o tailscale-healthcheck

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/tailscale-healthcheck /tailscale-healthcheck

ENTRYPOINT ["/tailscale-healthcheck"]
