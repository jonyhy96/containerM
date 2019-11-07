FROM golang:1.11 as builder

ENV GO111MODULE=on

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .
# Build the binary.
RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -ldflags="-w -s" -o main .


# final stage
FROM scratch
WORKDIR /containerM
COPY --from=builder /app /containerM/
ENTRYPOINT ["./main"]
