# Build stage
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with static linking
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o prometheus-docker-container-exporter .

# Final stage - scratch container
FROM scratch

# Copy the binary from builder stage
COPY --from=builder /app/prometheus-docker-container-exporter /prometheus-docker-container-exporter

# Expose port 8080
EXPOSE 8080

# Run the application
ENTRYPOINT ["/prometheus-docker-container-exporter"]