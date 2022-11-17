# build stage
FROM golang:1.19.3-alpine3.16 AS builder
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates
WORKDIR /build
COPY constants/ ./constants
COPY go.mod go.sum hub.go main.go client.go index.html ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o main .

# run stage
# scratch - empty container
FROM scratch AS preparer
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/main .
COPY --from=builder /build/index.html .
ENTRYPOINT ["/main"]
