FROM golang:1.22

WORKDIR ${GOPATH}/avito-shop/

COPY . ${GOPATH}/avito-shop/

RUN mkdir -p ./.bin

RUN go build -o ./.bin/avito_merch_store ./cmd/merch_store

EXPOSE 8080

CMD ["./.bin/avito_merch_store"]
