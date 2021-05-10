FROM fun.jfrog.io/dhub/govips:0.1.70 as builder

ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

# Set up build directories
RUN mkdir -p /app && \
    mkdir -p /BUILD && \
    mkdir -p /BUILD/db

# Build the goblog binary
COPY cmd /BUILD/cmd
COPY go.sum  /BUILD.go.sum
COPY go.mod /BUILD/go.mod
COPY internal /BUILD/internal
RUN cd /BUILD && go build -o /BUILD/dabloog cmd/goblog/main.go 

# Maxmind
FROM  debian:stretch-slim as maxmindupdate

#Should be set in environment
ARG ACCOUNT_ID="123"
ARG LICENSE_KEY="xxx"

RUN mkdir -p /app && \
    mkdir -p /BUILD && \
    mkdir -p /BUILD/db

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

# Production container
FROM fun.jfrog.io/dhub/vips:0.1.1

# Add user and set up temporary account
RUN mkdir /app && \
    mkdir app/temp && \
    addgroup app && \
    addgroup goblog && \
    adduser --home /app --system --no-create-home goblog goblog && \
    chown -R goblog:goblog /app && \
    chmod 1777 app/temp 

#Copy Stuff
COPY --from=builder /BUILD/dabloog /app/dabloog
COPY --from=maxmindupdate /BUILD/db /app/db
COPY sites /app/sites

WORKDIR /app

USER goblog
EXPOSE 5000
    
CMD ["./dabloog"]
