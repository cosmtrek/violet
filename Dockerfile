FROM golang:1.7

ADD . /go/src/github.com/cosmtrek/violet

WORKDIR /go/src/github.com/cosmtrek/violet

RUN make setup
RUN go install
CMD /go/bin/violet &

WORKDIR /go/src/github.com/cosmtrek/violet/example/tweet-search
RUN go install

EXPOSE 6060
EXPOSE 8100
ENTRYPOINT /go/bin/tweet-search