# Stage 1: Build Golang static binary
FROM golang:1.22-alpine AS builder

WORKDIR /app

# step-1 copy dependency module files and download dependencies
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# step-2 copy backend source code
COPY backend/ ./

# step-3 compile production binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

# Stage 2: Create lightweight production runtime image
FROM alpine:latest

# Install ca-certificates for secure HTTPS requests if needed
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# step-4 copy built binary from builder stage
COPY --from=builder /app/server .

# step-5 expose default port 8080
EXPOSE 8080

# step-6 set entrypoint command
CMD ["./server"]
