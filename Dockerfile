FROM debian:stretch-slim

RUN mkdir /app && \
    mkdir /build && \
    mkdir app/temp && \
    chmod 1777 app/temp && \
    groupadd -g 1001 goblog && \
    useradd -r -u 1001 -g goblog goblog && \
    chown -R goblog:goblog /app 

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y wget curl ca-certificates \
    build-essential automake pkg-config libexpat1-dev \
    libglib2.0-dev fftw3-dev liblcms2-dev libexif-dev libjpeg-dev

ENV LIBVIPS_VERSION_MAJOR 8
ENV LIBVIPS_VERSION_MINOR 9
ENV LIBVIPS_VERSION_PATCH 1
ENV LIBVIPS_VERSION $LIBVIPS_VERSION_MAJOR.$LIBVIPS_VERSION_MINOR.$LIBVIPS_VERSION_PATCH

  # download libvips

RUN \
  cd /tmp && \
  curl -L -O https://github.com/jcupitt/libvips/releases/download/v$LIBVIPS_VERSION/vips-$LIBVIPS_VERSION.tar.gz && \
  tar zxvf vips-$LIBVIPS_VERSION.tar.gz

# build libvips
RUN \
  cd /tmp/vips-$LIBVIPS_VERSION && \
  ./configure --without-gsf --without-orc --without-OpenEXR \
  --without-nifti --without-heif --without-rsvg \
  --without-openslide --without-matio --without-radiance \
  --without-libwebp --without-tiff --without-giflib \
  --without-imagequant --without-pangoft2 --enable-debug=no --without-python $1 && \
  make && \
  make install && \
  ldconfig  

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
  apt-get remove -y wget curl build-essential automake && \
  apt-get autoremove -y && \
  apt-get autoclean && \
  apt-get clean && \
  rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/* && \
  rm -rf /BUILD && \
  rm -rf /usr/local/go 

USER goblog
EXPOSE 5000
    
CMD ["./dabloog"] 