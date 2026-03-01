# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build
WORKDIR /app

RUN apk add --no-cache git

# ВАЖНО: чтобы не ходить в sum.golang.org (у тебя он рвется)
ENV GOPROXY=https://goproxy.io,direct
ENV GOSUMDB=off

COPY go.mod ./

# ретраи на скачивание модулей (сеть может отваливаться)
RUN for i in 1 2 3 4 5; do go mod download && break || (echo "retry $i" && sleep 2); done

COPY . .

# -mod=mod позволяет go дописать go.sum внутри контейнера и не падать
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=mod -o /bin/server ./cmd/server

FROM alpine:3.20
WORKDIR /app
RUN adduser -D -H appuser
USER appuser

ENV APP_PORT=8080
EXPOSE 8080

COPY --from=build /bin/server /app/server
CMD ["/app/server"]
