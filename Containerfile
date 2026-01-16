# Use the official Golang image as the base image
FROM docker.io/golang:1.24-alpine as builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files
COPY ./k8s-getsecret/* ./

# Download the dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o /app/k8s-getsecret .

# Switch to a smaller base image for the final stage
FROM scratch

# Set the working directory inside the container
WORKDIR /app

# Copy the built Go binary from the previous stage
COPY --from=builder /app/k8s-getsecret .

# Copy the kubeconfig file if needed (optional)
# COPY kubeconfig /root/.kube/config

# Run the application
CMD ["./k8s-getsecret"]
