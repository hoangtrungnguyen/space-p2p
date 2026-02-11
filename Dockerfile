# syntax=docker/dockerfile:1

# --- BUILD STAGE ---
ARG GO_VERSION=1.25.6
FROM golang:${GO_VERSION}-alpine AS build

# Install build dependencies
RUN apk add --no-cache ca-certificates git

WORKDIR /src

# Copy dependency files and download
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/server main.go

# --- RUN STAGE ---
FROM alpine:latest AS final

# Install ca-certificates for secure connections (required for LiveKit API)
RUN apk add --no-cache ca-certificates

# Create a non-privileged user for security
ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser

WORKDIR /app

# Copy the binary from the build stage
COPY --from=build /bin/server /app/server

# Set proper permissions
RUN chown -R appuser:appuser /app
USER appuser

# Expose the default port
EXPOSE 8080

# Run the application
# Note: Use environment variables or a mounted secret to provide LiveKit credentials at runtime
CMD ["/app/server"]
