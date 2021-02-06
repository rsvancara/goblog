#FROM debian:stretch-slim/ads as builder
FROM rsvancara/govips:0.1.16 as builder

RUN mkdir /app && \
    mkdir /BUILD 

# Build the goblog binary
COPY cmd /BUILD/cmd
COPY go.sum  /BUILD.go.sum
COPY go.mod /BUILD/go.mod
COPY vendor /BUILD/vendor
COPY internal /BUILD/internal
RUN cd /BUILD && PKG_CONFIG_PATH=$PKG_CONFIG_PATH:/opt/vips/lib/pkgconfig LD_LIBRARY_PATH=/opt/vips/lib /usr/local/go/bin/go build -o /BUILD/dabloog cmd/goblog/main.go 

FROM debian:stretch-slim

RUN mkdir /app && \
    groupadd -g 1001 goblog && \
    useradd -r -u 1001 -g goblog goblog && \
    chown -R goblog:goblog /app && \
    mkdir app/temp && \
    chmod 1777 app/temp 

COPY --from=builder /opt/vips /opt/vips
COPY --from=builder /BUILD/dabloog /app/dabloog

WORKDIR /app

# Copy the site directories
COPY sites sites

# Copy the database directories
COPY db db

RUN \
  apt-get update && \
  apt-get upgrade -y && \
  apt-get install -y libjpeg62 libexpat1 libglib2.0-0 libfftw3-3 liblcms2-2 libexif12 ca-certificates && \
  apt-get clean

USER goblog
EXPOSE 5000

ENV LD_LIBRARY_PATH=/opt/vips/lib
    
CMD ["./dabloog"]
