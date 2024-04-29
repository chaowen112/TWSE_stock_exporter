FROM golang:1.22.1
COPY . /go/src/
WORKDIR /go/src/
RUN go build -o /go/bin/app
CMD ["/go/bin/app"]
