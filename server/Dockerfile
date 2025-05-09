# Start from official Go image
FROM golang:1.22.2 AS builder

WORKDIR /app
COPY . .

# Enable Go modules and download dependencies
RUN go mod tidy
RUN go build -o server .

# Use a minimal base image for production
FROM debian:bookworm-slim

# Install CA certificates for HTTPS support
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /root/
COPY --from=builder /app/server .
COPY chatterCount.json .

# Expose the port used by your server
EXPOSE 9002

# Start the server
CMD ["./server"]
