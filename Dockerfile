FROM tryingadventure.jfrog.io/dhub/govips:0.1.58 as builder

ENV PATH /usr/local/go/bin:$PATH

ENV GOLANG_VERSION 1.15.8

RUN apk add --no-cache \
		ca-certificates

RUN apk add --no-cache --virtual .build-dependencies autoconf automake build-base pkgconfig cmake gobject-introspection \
    git libtool nasm zlib-dev libxml2-dev libxslt-dev glib-dev \
    libexif-dev lcms2-dev fftw-dev giflib-dev libpng-dev libwebp-dev orc-dev tiff-dev \
    poppler-dev librsvg-dev wget libheif-dev gtk-doc 

# set up nsswitch.conf for Go's "netgo" implementation
# - https://github.com/golang/go/blob/go1.9.1/src/net/conf.go#L194-L275
# - docker run --rm debian:stretch grep '^hosts:' /etc/nsswitch.conf
RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf

RUN set -eux; \
	apk add --no-cache --virtual .build-deps \
		bash \
		gcc \
		gnupg \
		go \
		musl-dev \
		openssl \
	; \
	export \
# set GOROOT_BOOTSTRAP such that we can actually build Go
		GOROOT_BOOTSTRAP="$(go env GOROOT)" \
# ... and set "cross-building" related vars to the installed system's values so that we create a build targeting the proper arch
# (for example, if our build host is GOARCH=amd64, but our build env/image is GOARCH=386, our build needs GOARCH=386)
		GOOS="$(go env GOOS)" \
		GOARCH="$(go env GOARCH)" \
		GOHOSTOS="$(go env GOHOSTOS)" \
		GOHOSTARCH="$(go env GOHOSTARCH)" \
	; \
# also explicitly set GO386 and GOARM if appropriate
# https://github.com/docker-library/golang/issues/184
	apkArch="$(apk --print-arch)"; \
	case "$apkArch" in \
		armhf) export GOARM='6' ;; \
		armv7) export GOARM='7' ;; \
		x86) export GO386='387' ;; \
	esac; \
	\
# https://github.com/golang/go/issues/38536#issuecomment-616897960
	url='https://storage.googleapis.com/golang/go1.15.8.src.tar.gz'; \
	sha256='540c0ab7781084d124991321ed1458e479982de94454a98afab6acadf38497c2'; \
	\
	wget -O go.tgz.asc "$url.asc"; \
	wget -O go.tgz "$url"; \
	echo "$sha256 *go.tgz" | sha256sum -c -; \
	\
# https://github.com/golang/go/issues/14739#issuecomment-324767697
	export GNUPGHOME="$(mktemp -d)"; \
# https://www.google.com/linuxrepositories/
	gpg --batch --keyserver ha.pool.sks-keyservers.net --recv-keys 'EB4C 1BFD 4F04 2F6D DDCC EC91 7721 F63B D38B 4796'; \
	gpg --batch --verify go.tgz.asc go.tgz; \
	gpgconf --kill all; \
	rm -rf "$GNUPGHOME" go.tgz.asc; \
	\
	tar -C /usr/local -xzf go.tgz; \
	rm go.tgz; \
	\
	goEnv="$(go env | sed -rn -e '/^GO(OS|ARCH|ARM|386)=/s//export \0/p')"; \
	eval "$goEnv"; \
	[ -n "$GOOS" ]; \
	[ -n "$GOARCH" ]; \
	( \
		cd /usr/local/go/src; \
		./make.bash; \
	); \
	\
	apk del --no-network .build-deps; \
	\
# pre-compile the standard library, just like the official binary release tarballs do
	go install std; \
# go install: -race is only supported on linux/amd64, linux/ppc64le, linux/arm64, freebsd/amd64, netbsd/amd64, darwin/amd64 and windows/amd64
#	go install -race std; \
	\
# remove a few intermediate / bootstrapping files the official binary release tarballs do not contain
	rm -rf \
		/usr/local/go/pkg/*/cmd \
		/usr/local/go/pkg/bootstrap \
		/usr/local/go/pkg/obj \
		/usr/local/go/pkg/tool/*/api \
		/usr/local/go/pkg/tool/*/go_bootstrap \
		/usr/local/go/src/cmd/dist/dist \
	; \
	\
	go version

ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH
RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

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

# Productioncontainer
FROM tryingadventure.jfrog.io/dhub/govips:0.1.58

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
