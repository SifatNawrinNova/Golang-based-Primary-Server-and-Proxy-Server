# Use the official Golang image to create a build artifact.
FROM golang:1.21.3 as builder

# Copy the local package files to the container's workspace.
ADD . /go/src/server

# Build the command inside the container.
WORKDIR /go/src/server
RUN go build -o /server

# Final stage
FROM golang:1.21.3
COPY --from=builder /server /server


# Run the server binary.
ENTRYPOINT ["/server"]
