# Build stage
FROM golang:1.23.1-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy
# Copy whole project and build the server
COPY . .
RUN go build -ldflags="-s -w" -o server main.go

# Runtime stage to reduce image size
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server .
COPY --from=builder /app/static ./static
EXPOSE 3000
CMD ["./server"]