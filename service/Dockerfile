FROM golang:1.13.1 AS builder

ENV GO111MODULE off

WORKDIR /go/src/github.com/Tahler/isotope/service

COPY . .
RUN go get -v

RUN CGO_ENABLED=0 GOOS=linux \
    go build -a -installsuffix cgo -o ./main ./main.go

FROM scratch
COPY --from=builder \
    /go/src/github.com/Tahler/isotope/service/main /usr/local/bin/service

EXPOSE 8080 8081
ENTRYPOINT [ "/usr/local/bin/service" ]
