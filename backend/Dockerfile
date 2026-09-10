FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

RUN if [ -d "./backend" ]; then \
        cd ./backend && go mod download && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server; \
    else \
        go mod download && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server; \
    fi

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/
COPY --from=builder /app/server .

EXPOSE 8080
ENV PORT=8080
ENV DB_PATH=/root/sequencer.db

CMD ["./server"]
