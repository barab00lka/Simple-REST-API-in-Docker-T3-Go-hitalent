# Use a multi-stage build to keep the final image small
# Stage 1: Build the Go application
FROM golang:alpine AS builder

# Set the Current Working Directory to /app
WORKDIR /app
# Copy go mod and sum files
COPY ./src/go.mod ./src/go.sum ./
# Download all dependencies
RUN go mod download
# Copy the source from the current directory to the Working Directory
COPY ./src .
# Build the Go application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Stage 2: Test image – contains full source + dependencies (from builder)
FROM golang:alpine AS test
WORKDIR /app
# Copy the Go module cache from builder (saves downloading again)
COPY --from=builder /go/pkg/mod /go/pkg/mod
# Copy the entire source code (including tests, migrations, internal packages)
COPY --from=builder /app .
CMD ["./main"]

# Stage 3: Create a minimal alpine production image for the application
FROM alpine:latest
# Set the Current Working Directory to /app
WORKDIR /app
# Copy the binary from the builder
COPY --from=builder /app/main .
# Command to run the executable
CMD ["./main"]
