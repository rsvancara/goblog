FROM debian:stretch-slim as builder

RUN mkdir /app && \
    mkdir /BUILD 

RUN apt-get update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y wget curl ca-certificates \
    build-essential automake pkg-config libexpat1-dev \
    libglib2.0-dev fftw3-dev liblcms2-dev libexif-dev libjpeg-dev libwebp-dev

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
  ./configure --prefix=/opt/vips --with-pkg-config -without-gsf --without-orc --without-OpenEXR \
  --without-nifti --without-heif --without-rsvg \
  --without-openslide --without-matio --without-radiance \
  --without-tiff --without-giflib \
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
