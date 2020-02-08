FROM golang:1.13.4-stretch

RUN apt-get update && \
    apt-get install -y libvips-dev && \
    apt-get clean 

RUN mkdir BUILD
WORKDIR /BUILD

# Build the binary
COPY blog.go /BUILD/blog.go
COPY go.sum  /BUILD.go.sum
COPY go.mod /BUILD/go.mod
COPY vendor /BUILD/vendor
COPY blog /BUILD/blog

RUN go build -o dabloog blog.go 

FROM debian:stretch-slim

RUN apt-get update &&  \
    apt-get install -y libvips ca-certificates && \
    apt-get clean 

RUN mkdir app && \
    mkdir app/temp && \
    chmod 1777 app/temp && \
    groupadd -g 1001 goblog && \
    useradd -r -u 1001 -g goblog goblog && \
    chown -R goblog:goblog /app 

WORKDIR /app

COPY --from=0 /BUILD/dabloog .

# Copy the site directories
COPY visualintrigue.com visualintrigue.com
COPY tinytrailerfun.com tinytrailerfun.com
COPY db db

USER goblog
EXPOSE 5000
    
CMD ["./dabloog"] 
