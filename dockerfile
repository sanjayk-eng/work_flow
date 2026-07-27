FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN go build -o app ./cmd/api

# -----------------------------
# Runtime image
# -----------------------------
FROM alpine:latest

WORKDIR /root/

# Copy the compiled binary
COPY --from=builder /app/app .

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./app"]