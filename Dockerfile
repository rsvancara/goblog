FROM golang:1.13.4-stretch

RUN mkdir BUILD
WORKDIR /BUILD

COPY blog.go /BUILD/blog.go
COPY go.sum  /BUILD.go.sum
COPY go.mod /BUILD/go.mod
COPY vendor /BUILD/vendor
COPY blog /BUILD/blog

#RUN go mod download && CGO_ENABLED=1 GOOS=linux go build blog.go 

RUN CGO_ENABLED=1 GOOS=linux go build -o dabloog blog.go 

FROM ubuntu:18.04

RUN apt-get update &&  \
    apt-get clean && \
    mkdir app

WORKDIR /app

COPY --from=0 /BUILD/dabloog .

COPY templates templates
COPY static static
    
CMD ["./dabloog"] 
