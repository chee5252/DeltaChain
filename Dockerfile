# Build Gdch in a stock Go builder container
FROM golang:1.11-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers

ADD . /go-deltachaineum
RUN cd /go-deltachaineum && make gdch

# Pull Gdch into a second stage deploy alpine container
FROM alpine:latest

RUN apk add --no-cache ca-certificates
COPY --from=builder /go-deltachaineum/build/bin/gdch /usr/local/bin/

EXPOSE 8545 8546 30303 30303/udp
ENTRYPOINT ["gdch"]
