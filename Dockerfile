FROM golang:1.24

RUN go install github.com/mitranim/gow@latest

WORKDIR /opt/app-root/src
COPY . /opt/app-root/src/

CMD [ "gow", "run", "tournament-manager/cmd" ]
