FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o hf-local .

FROM alpine:latest

RUN apk add --no-cache sqlite-libs

WORKDIR /app

COPY --from=builder /app/hf-local /app/hf-local

ENV HF_LOCAL_PORT=8080
ENV HF_LOCAL_DATA_DIR=/app/data

EXPOSE 8080

CMD ["./hf-local"]
