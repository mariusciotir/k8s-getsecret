# Use the official Golang image as the base image
FROM docker.io/golang:1.24-alpine as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY ./k8s-getsecret/* ./

# Download the dependencies
RUN go mod download

# Build the Go application
RUN CGO_ENABLED=0 go build -v -o k8s-getsecret main.go

# Switch to a smaller base image for the final stage
FROM scratch

# Copy the built Go binary from the previous stage
COPY --from=builder /app/k8s-getsecret /k8s-getsecret

# Run the application
ENTRYPOINT ["/k8s-getsecret"]
