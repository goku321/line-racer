FROM golang

ADD . /go/src/github.com/goku321/line-racer

RUN go install github.com/goku321/line-racer

EXPOSE 3000

ENTRYPOINT [ ./go/bin/line-racer ]