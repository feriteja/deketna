# Stage 1: Build Stage (Production Binary Build)
FROM golang:1.22 AS builder

WORKDIR /app

# Copy dependencies and build
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o main .

# Stage 2: Runtime Stage (For Development & Production)
FROM golang:1.22

WORKDIR /app

# Copy binary from build stage for production
COPY --from=builder /app/main .

# Set default command for development
CMD [ "go", "run", "."]

# Expose app port
EXPOSE 8080
