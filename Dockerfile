# Build Gman in a stock Go builder container #shang and yang
FROM golang:1.9-alpine 

RUN apk add --no-cache make gcc musl-dev linux-headers

ADD ./MATRIX-TESTNET /go-matrix

RUN cd /go-matrix && make gman

# Pull Gman into a second stage deploy alpine container

RUN apk add --no-cache ca-certificates

RUN  ln -s  /go-matrix/build/bin/gman /usr/local/bin/gman

#EXPOSE 8545 8546 30303 30303/udp 30304/udp

EXPOSE	40404 50505 8341

ENTRYPOINT ["gman"]
