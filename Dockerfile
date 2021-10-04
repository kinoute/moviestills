# build stage
FROM golang:1.17.1-alpine as builder

RUN apk update && \
    apk --no-cache add ca-certificates=20191127-r5

WORKDIR /app

COPY ./ .

RUN go mod download && \
    GOOS=linux go build -ldflags "-s -w"

# final stage
FROM alpine:3.14

WORKDIR /root

COPY --from=builder /app/moviestills .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 80

CMD ["./moviestills"]
