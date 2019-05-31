FROM golang:1.11-alpine3.8 As base
# Install git.
# Git is required for fetching the dependencies.
WORKDIR $GOPATH/src/containerM
COPY . .
# Build the binary.
RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -ldflags="-w -s" -o main .


FROM scratch
# Copy our static executable.
ENV GOPATH /go
COPY --from=base /go/src/containerM /go/src/containerM
# Run the hello binary.
ENTRYPOINT ["/go/src/containerM/main"]