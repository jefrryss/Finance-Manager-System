FROM golang:1.25.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .

RUN swag init -g cmd/main.go

RUN CGO_ENABLED=0 GOOS=linux go build -o finance_manager ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/finance_manager .

COPY --from=builder /app/configs ./configs

 COPY --from=builder /app/.env .

EXPOSE 8080

CMD ["./finance_manager"]