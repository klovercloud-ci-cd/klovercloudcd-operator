# Build the manager binary
FROM golang:1.16  as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

Copy . .
COPY controllers/ controllers/
# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM debian:buster-slim
WORKDIR /
COPY --from=builder /workspace/manager .
COPY --from=builder /workspace/controllers .
USER 65532:65532

ENTRYPOINT ["/manager"]

#FROM golang:1.16 as builder
#RUN apt-get update && apt-get install -y nocache git ca-certificates && update-ca-certificates
#WORKDIR /app
#COPY go.mod go.sum ./
##RUN go env -w GOPROXY="https://goproxy.io,direct"
#RUN go mod download
#COPY . .
#RUN go build -o /app/bin/manager .
#
#
#
#FROM debian:buster-slim
#COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
#WORKDIR /app
#COPY --from=builder /app/bin /app
## Run the executable
#CMD ["./manager"]