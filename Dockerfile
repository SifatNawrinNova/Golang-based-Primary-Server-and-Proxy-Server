# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
FROM golang:1.18 as builder

# Copy the local package files to the container's workspace.
ADD . /go/src/app

# Build the command inside the container.
WORKDIR /go/src/app
RUN go build -o /app

# Use a Docker multi-stage build to create a lean image.
FROM debian:buster-slim
COPY --from=builder /app /app

# Run the binary program produced by `go install`.
CMD ["/app"]
