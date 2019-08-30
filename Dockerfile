# Build the manager binary
FROM golang:1.10.3 as builder

# Copy in the go src
WORKDIR /go/src/github.com/triggermesh/vsphere-source
COPY cmd/    cmd/
COPY vendor/ vendor/
COPY pkg/    pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager github.com/triggermesh/vsphere-source/cmd/manager

# Copy the controller-manager into a thin image
FROM ubuntu:latest
WORKDIR /
COPY --from=builder /go/src/github.com/triggermesh/vsphere-source/manager .
ENTRYPOINT ["/manager"]
