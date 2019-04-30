# Build Gman in a stock Go builder container
FROM golang:1.9-alpine as builder

RUN apk add --no-cache  make gcc musl-dev linux-headers git

ADD . /go-matrix

RUN cd /go-matrix &&  chmod +x Start && chmod +x build/env.sh && make gman

# Pull Gman into a second stage deploy alpine container

FROM alpine:latest

RUN apk add --no-cache tmux && mkdir /root/.matrix/ -p 

COPY --from=builder /go-matrix/build/bin/gman /usr/bin/

COPY --from=builder /go-matrix/Start  /usr/bin/

COPY --from=builder /go-matrix/MANGenesis.json /root/.matrix/

COPY --from=builder /go-matrix/man.json  /root/.matrix/

WORKDIR /root/.matrix/

EXPOSE 8341 50505 50505/udp

CMD ["Start"]
