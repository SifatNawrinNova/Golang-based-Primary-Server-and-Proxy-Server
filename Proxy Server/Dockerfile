# Use the official Golang image to create a build artifact.
FROM golang:1.21.3 as builder

# Copy the local package files to the container's workspace.
ADD . /go/src/proxy

# Build the command inside the container.
WORKDIR /go/src/proxy
RUN go build -o /proxy

# Final stage
FROM golang:1.21.3
COPY --from=builder /proxy /proxy

# Set the client binary as the entrypoint.
ENTRYPOINT ["/proxy"]
