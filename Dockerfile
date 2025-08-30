# Step 1: Build the Go binary
FROM golang:1.24 AS builder

# Set the working directory inside the container for the Go app
WORKDIR /app

# Copy the Go module files for caching dependencies
COPY go.mod ./

# Download Go dependencies
RUN go mod tidy

# Copy the entire Go source code into the container
COPY . .

# Build the Go binary
RUN GOOS=linux GOARCH=amd64 go build -o myapp .

# Step 2: Prepare environment for the CLI and the Go app
FROM node:22-alpine AS cli_installer

# Install necessary dependencies for the CLI
RUN apk --no-cache add curl git

# Install the required CLI tool using npm   
RUN npm install -g @bitwarden/cli --legacy-peer-deps

# # Step 3: Create a minimal runtime image with both Go and CLI installed
# FROM alpine:latest

# # Install dependencies for running the Go app and CLI
# RUN apk --no-cache add ca-certificates bash nodejs npm git

# Copy the Go binary from the builder stage
COPY --from=builder /app/myapp /root/myapp

# # Copy the installed CLI from the cli_installer stage
# COPY --from=cli_installer /usr/local/bin/bw /usr/local/bin/

# Set the working directory
WORKDIR /root/

# Command to run the Go binary
CMD ["./myapp"]
