FROM golang:1.18 as build

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
RUN go get github.com/vektra/mockery/.../
#Adding changed files last for hitting docker layer cache
COPY . .
RUN go generate ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /app/bee-search

# Switch to a small base image
FROM scratch

# Get the TLS CA certificates from the build container, they're not provided by busybox.
COPY --from=build /etc/ssl/certs /etc/ssl/certs

# copy app to bin directory, and set it as entrypoint
WORKDIR /app
COPY --from=build /app/bee-search /app/bee-search
COPY --from=build /app/resources/ /app/resources

EXPOSE 8082

ENTRYPOINT ["/app/bee-search"]
