FROM golang

ADD . /go/src/github.com/goku321/line-racer

RUN go install github.com/goku321/line-racer

ENTRYPOINT [ "/go/src/github.com/goku321/line-racer/init-race.sh" ]