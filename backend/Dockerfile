FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN if [ -f go.sum ]; then cp go.sum ./; fi
RUN go mod download
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations

EXPOSE 5000

CMD ["./main"]
