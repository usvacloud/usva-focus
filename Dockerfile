FROM golang:1.17-alpine3.15 as builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY *.go .
RUN go build main.go

FROM alpine:3.15
RUN apk add --no-cache \
  bash redis

WORKDIR /app
COPY --from=builder /build/main ./usvad
COPY views views/
COPY entrypoint.sh .
ENV USVA_SEEDS=fiesta.usva.io
ENV GIN_MODE=release
ENTRYPOINT [ "/app/entrypoint.sh" ]