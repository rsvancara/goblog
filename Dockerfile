#FROM debian:stretch-slim/ads as builder
FROM rsvancara/govips:0.1.16 as builder

RUN mkdir -p /app && \
    mkdir -p /BUILD && \
    mkdir -p /BUILD/db

# Build the goblog binary
COPY cmd /BUILD/cmd
COPY go.sum  /BUILD.go.sum
COPY go.mod /BUILD/go.mod
COPY internal /BUILD/internal
RUN cd /BUILD && PKG_CONFIG_PATH=$PKG_CONFIG_PATH:/opt/vips/lib/pkgconfig LD_LIBRARY_PATH=/opt/vips/lib /usr/local/go/bin/go build -o /BUILD/dabloog cmd/goblog/main.go 

# Maxmind
FROM  debian:stretch-slim as maxmindupdate

#Shouldbe set in environment
ARG ACCOUNT_ID="123"
ARG LICENSE_KEY="xxx"

RUN echo $ACCOUNT_ID

RUN \
  apt-get update && \
  apt-get upgrade -y && \
  apt-get install -y wget ca-certificates && \
  apt-get clean

WORKDIR /BUILD

RUN wget https://github.com/maxmind/geoipupdate/releases/download/v4.6.0/geoipupdate_4.6.0_linux_amd64.deb && \
    dpkg -i geoipupdate_4.6.0_linux_amd64.deb 

RUN echo "AccountID ${ACCOUNT_ID}" > /etc/GeoIP.conf && \
    echo "LicenseKey ${LICENSE_KEY}" >> /etc/GeoIP.conf && \
    echo "EditionIDs GeoIP2-City GeoIP2-Country GeoLite2-ASN GeoLite2-City GeoLite2-Country" >> /etc/GeoIP.conf && \
    echo "DatabaseDirectory /BUILD/db" >> /etc/GeoIP.conf && \
    /usr/bin/geoipupdate -v

# Productioncontainer
FROM debian:stretch-slim

RUN \
  apt-get update && \
  apt-get upgrade -y && \
  apt-get install -y libjpeg62 libexpat1 libglib2.0-0 libfftw3-3 liblcms2-2 libexif12 ca-certificates && \
  apt-get clean

RUN mkdir /app && \
    groupadd -g 1001 goblog && \
    useradd -r -u 1001 -g goblog goblog && \
    chown -R goblog:goblog /app && \
    mkdir app/temp && \
    chmod 1777 app/temp 

COPY --from=builder /opt/vips /opt/vips
COPY --from=builder /BUILD/dabloog /app/dabloog
COPY --from=maxmindupdate /BUILD/db /app/db
COPY sites /app/sites

WORKDIR /app

USER goblog
EXPOSE 5000

ENV LD_LIBRARY_PATH=/opt/vips/lib
    
CMD ["./dabloog"]