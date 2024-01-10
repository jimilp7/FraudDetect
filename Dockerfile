FROM golang:1.21

# Set the working directory in the container
WORKDIR /FraudDetection

# Copy the local package files to the container's workspace
COPY . .

# Download all dependencies
RUN go mod download

# Build the Go app
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Run the executable
CMD ["./main"]