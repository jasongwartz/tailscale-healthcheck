FROM golang AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go .
RUN CGO_ENABLED=0 GOOS=linux \
    go build -o tailscale-healthcheck
RUN mkdir /tsnet-state

FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/tailscale-healthcheck /tailscale-healthcheck

# 65532 is the user "nonroot" in the distroless image, as per:
# https://github.com/GoogleContainerTools/distroless/issues/427#issuecomment-547874186
COPY --from=builder --chown=65532:65532 /tsnet-state /tsnet-state
ENV TSNET_STATE_DIR="/tsnet-state"
USER nonroot

ENTRYPOINT ["/tailscale-healthcheck"]
