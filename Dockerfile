FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR $GOPATH/src/jeli.cc/go-gnudip/
COPY . .

RUN go get -d -v

RUN go build -o /go/bin/go-gnudip

FROM scratch

COPY --from=builder /go/bin/go-gnudip /go/bin/go-gnudip

ENTRYPOINT ["/go/bin/go-gnudip"]
