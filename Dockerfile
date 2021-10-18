FROM golang:1.17-alpine

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Copy the code into the container
COPY . .
RUN go mod download

# Build the application
RUN go build -o main .

# Move to /dist directory as the place for resulting binary folder
WORKDIR /dist

# Copy binary from build to main folder
RUN cp /build/main . && \
    rm -rf /build

# Export necessary port
EXPOSE 6081

# Command to run when starting the container
CMD ["/dist/main"]
