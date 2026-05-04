FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN go install github.com/shinagawa-web/gomarklint/v2@latest

FROM alpine:latest
RUN apk --no-cache add ca-certificates curl jq

COPY --from=builder /go/bin/gomarklint /usr/local/bin/gomarklint
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
