FROM  golang:1.17.6-alpine3.15 as builder

#ENV GOPATH /usr/local/go
#ENV PATH $GOPATH/bin:$PATH

# Set up build directories
RUN mkdir -p /app && \
    mkdir -p /BUILD && \
    mkdir -p /BUILD/db

# Build the goblog binary
COPY cmd /BUILD/cmd
COPY go.sum  /BUILD/go.sum
COPY go.mod /BUILD/go.mod
COPY internal /BUILD/internal
RUN cd /BUILD && go mod vendor && go mod download
RUN cd /BUILD && go build -o /BUILD/dabloog cmd/goblog/main.go 



# Production container
FROM alpine

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
COPY sites /app/sites

WORKDIR /app

USER goblog
EXPOSE 5000
    
CMD ["./dabloog"]
