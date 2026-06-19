# ----------------------------------------------------------------------
#  STAGE 1: Build the Go Binary
# ----------------------------------------------------------------------
FROM golang:1.26-alpine AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy dependency manifests first (improves Docker layer caching speeds)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Compile the binary with optimization flags turned off for debugging, 
# static linking, and stripping out unused debug symbols to save space.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o identity-service ./cmd/api/main.go

# ----------------------------------------------------------------------
#  STAGE 2: Run the Small Binary
# ----------------------------------------------------------------------
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates just in case your app makes outbound HTTPS external calls (like SMS APIs)
RUN apk --no-cache add ca-certificates

# Copy the compiled binary over from the builder stage
COPY --from=builder /app/identity-service .

# Expose the application port
EXPOSE 8080

# Execute the application
CMD ["./identity-service"]