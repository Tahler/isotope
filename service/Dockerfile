FROM golang:1.10.2 AS builder

RUN go get -u github.com/golang/dep/cmd/dep

WORKDIR /go/src/github.com/Tahler/isotope/service

COPY . .
RUN dep ensure -vendor-only

RUN CGO_ENABLED=0 GOOS=linux \
    go build -a -installsuffix cgo -o ./main ./main.go

FROM scratch
COPY --from=builder \
    /go/src/github.com/Tahler/isotope/service/main /usr/local/bin/service

EXPOSE 8080 8081
ENTRYPOINT ["/usr/local/bin/service"]
