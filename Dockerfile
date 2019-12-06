FROM golang:1.13.4-stretch

RUN mkdir /build

WORKDIR /build

COPY blog.go .
COPY go.sum  .
COPY go.mod .
COPY vendor .

RUN go mod download && CGO_ENABLED=1 GOOS=linux go build blog.go 

FROM ubuntu:18.04

RUN apt-get update &&  \
    apt-get clean && \
    mkdir app
    
WORKDIR /app

COPY --from=0 /build/blog .

COPY templates templates

COPY nmap.db .

CMD ["./blog"] 
