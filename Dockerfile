FROM golang:1.13.1 AS builder

ENV GO111MODULE on

WORKDIR /go/src/github.com/Tahler/isotope

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux \
    go build -a -installsuffix cgo -o ./main service/main.go

FROM scratch
COPY --from=builder \
    /go/src/github.com/Tahler/isotope/main /usr/local/bin/service

EXPOSE 8080 8081
ENTRYPOINT [ "/usr/local/bin/service" ]
