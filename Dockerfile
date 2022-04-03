# base stage
FROM golang:1.18-alpine as base

RUN apk update && \
    apk --no-cache add \
    ca-certificates=20211220-r0 \
    gcc=10.3.1_git20211027-r0 \
    musl-dev=1.2.2-r7

WORKDIR /app

# Copy go mod and sum files
COPY ./go.mod ./go.sum ./

# Download all required packages
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY ./ .

# build stage
FROM base as builder

RUN GOOS=linux go build -ldflags "-s -w"

# final stage to serve binary
FROM alpine:3.15.3

WORKDIR /app

COPY --from=builder /app/moviestills .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["./moviestills"]
