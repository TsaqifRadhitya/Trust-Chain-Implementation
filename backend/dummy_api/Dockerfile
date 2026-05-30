# Build stage
FROM golang:alpine AS builder

WORKDIR /app

# Copy go mod file
COPY go.mod ./

# Download dependencies (if any)
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the application statically
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dummy_api .

# Run stage
FROM alpine:latest  

# Add ca-certificates in case you need to make HTTPS requests later
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/dummy_api .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./dummy_api"]
