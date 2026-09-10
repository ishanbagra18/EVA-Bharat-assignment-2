FROM golang:1.24-alpine AS builder

WORKDIR /app

# Support building from either repo root or backend context
COPY go.mod* go.sum* backend/go.mod* backend/go.sum* ./
RUN go mod download

COPY . .
RUN if [ -d "./cmd/server" ]; then \
        CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./cmd/server; \
    else \
        CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o server ./backend/cmd/server; \
    fi

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/
COPY --from=builder /app/server .

EXPOSE 8080
ENV PORT=8080
ENV DB_PATH=/root/sequencer.db

CMD ["./server"]
