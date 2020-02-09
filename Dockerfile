FROM debian:stretch-slim

RUN mkdir /app && \
    mkdir /build && \
    mkdir app/temp && \
    chmod 1777 app/temp && \
    groupadd -g 1001 goblog && \
    useradd -r -u 1001 -g goblog goblog && \
    chown -R goblog:goblog /app 

RUN apt-get update && \
    apt-get install -y wget ca-certificates libvips-dev

# Install golang
RUN wget https://dl.google.com/go/go1.13.4.linux-amd64.tar.gz && \
    tar -zxvf go1.13.4.linux-amd64.tar.gz && \
    rm -rf go1.13.4.linux-amd64.tar.gz && \
    mv go /usr/local/

# Build the binary
COPY blog.go /BUILD/blog.go
COPY go.sum  /BUILD.go.sum
COPY go.mod /BUILD/go.mod
COPY vendor /BUILD/vendor
COPY blog /BUILD/blog
RUN cd /BUILD && /usr/local/go/bin/go build -o /app/dabloog blog.go 

WORKDIR /app

# Copy the site directories
COPY visualintrigue.com visualintrigue.com
COPY tinytrailerfun.com tinytrailerfun.com
COPY db db

RUN \
  apt-get autoremove -y && \
  apt-get autoclean && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* && \
  rm -rf /usr/local/go 

USER goblog
EXPOSE 5000
    
CMD ["./dabloog"] 