# Stage 1: build web app (then discard this container)
FROM node:20-alpine AS web
WORKDIR /app
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

# Stage 2: build Go server
FROM golang:1.22.2 AS server-builder
WORKDIR /app
COPY server/ ./
RUN go mod tidy && go build -o server .

# Stage 3: single runnable image
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=server-builder /app/server .
COPY --from=web /app/dist ./web
COPY server/chatterCount.json .
ENV STATIC_DIR=/app/web
ENV PORT=8080
EXPOSE 8080
CMD ["./server"]
