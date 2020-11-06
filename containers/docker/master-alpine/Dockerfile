FROM alpine:3.7

RUN \
  apk add --update go git make gcc musl-dev linux-headers ca-certificates && \
  git clone --depth 1 --branch release/1.8 https://github.com/deltachaineum/go-deltachaineum && \
  (cd go-deltachaineum && make gdch) && \
  cp go-deltachaineum/build/bin/gdch /gdch && \
  apk del go git make gcc musl-dev linux-headers && \
  rm -rf /go-deltachaineum && rm -rf /var/cache/apk/*

EXPOSE 8545
EXPOSE 30303

ENTRYPOINT ["/gdch"]
