FROM golang:1.22

WORKDIR /app

COPY . .

RUN go build -o ./.bin/avito_merch_store ./cmd/merch_store

EXPOSE 8090

CMD ["./.bin/avito_merch_store"]
