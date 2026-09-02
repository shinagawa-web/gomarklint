FROM golang:1.27-alpine@sha256:26402d86be3d72e6a9410afa0108f03529f51f0c1b5eb7f503d0bc44cc7857ac AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o gomarklint .

FROM alpine:latest@sha256:28bd5fe8b56d1bd048e5babf5b10710ebe0bae67db86916198a6eec434943f8b
RUN apk --no-cache add ca-certificates curl jq

COPY --from=builder /app/gomarklint /usr/local/bin/gomarklint
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
