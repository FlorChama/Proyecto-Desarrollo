FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o ticketek-backend .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/ticketek-backend .
EXPOSE 8080
CMD ["./ticketek-backend"]
