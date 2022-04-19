FROM golang:1.17-alpine3.15 as builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY cmd cmd
COPY pkg pkg
RUN go build cmd/usva/usva.go

FROM alpine:3.15
RUN apk add --no-cache \
  bash redis

WORKDIR /app
COPY --from=builder /build/usva ./usva
COPY templates templates/
COPY entrypoint.sh .

ENV GIN_MODE=release
ENTRYPOINT [ "/app/entrypoint.sh" ]