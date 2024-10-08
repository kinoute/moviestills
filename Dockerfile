# base stage
FROM golang:1.23.1-alpine3.19 as base

RUN apk update && \
    apk --no-cache add \
    ca-certificates~=20240226-r0 \
    gcc~=13.2.1_git20231014-r0 \
    musl-dev~=1.2.4_git20230717-r4

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
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/moviestills .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["./moviestills"]
