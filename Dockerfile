FROM golang:1.7

ADD . /go/src/github.com/cosmtrek/violet

WORKDIR /go/src/github.com/cosmtrek/violet/example/tweets-search
RUN go install

WORKDIR /go/src/github.com/cosmtrek/violet
RUN make setup
RUN go install

EXPOSE 6060
EXPOSE 8100

COPY docker-entrypoint.sh /usr/local/bin/
ENTRYPOINT ["docker-entrypoint.sh"]
