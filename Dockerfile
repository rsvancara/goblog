FROM golang:1.13.4-stretch

RUN mkdir BUILD
WORKDIR /BUILD

# Build the binary
COPY blog.go /BUILD/blog.go
COPY go.sum  /BUILD.go.sum
COPY go.mod /BUILD/go.mod
COPY vendor /BUILD/vendor
COPY blog /BUILD/blog

RUN CGO_ENABLED=1 GOOS=linux go build -o dabloog blog.go 

FROM ubuntu:18.04

RUN apt-get update &&  \
    apt-get clean && \
    mkdir app

WORKDIR /app

COPY --from=0 /BUILD/dabloog .

# Copy the site directories
COPY visualintrigue.com visualintrigue.com
COPY tinycamperfun.com tinycamperfun.com

EXPOSE 5000
    
CMD ["./dabloog"] 
