# GoBlog - Go Based Photography Blogging Software

This is my take on a simple photography blog platform that I am using for my blog sites. 

Note: This software is a work in progress.  You can see it action here:

https://tryingadventure.com

https://diytinytrailer.com

## Goals

1.  Write it in golang, yes I wanted to learn Go and this was a good way to do it.
2.  Make it work for may blog platform.  Basically there is an engine and you can put it into any chassis you want and make it go!
3.  Run in a container environment (12 factor applications)
4.  Support advanced filtering for managing content display using GeoIP (requires maxmind geoip database...you can use the light edition for free)
5.  Use wiki markdown language for easy content creation.  I wanted to avoide WSISYG editors as I have found they are a pain in the ass to manage in terms of stylesheets. 

# Instructions

Running this on MacOS, you need to install libvips for the image processing library:

```bash
brew install libvips
```

Install Mongodb

```bash
brew tap mongodb/brew
brew install mongodb-community@4.4
```

Install Redis

```bash
brew update
brew install redis
```

Install the maxmind geoip databases.  This is needed for tracking users and session reports

Once the prequisites are installed, create a script below.  Substitute your redis, mongodb, and S3 credentials for the application.  



```bash
CGO_CFLAGS_ALLOW=-Xpreprocessor \
AWS_PROFILE="xxxx" \
S3_BUCKET="xxxx" \
AWS_ACCESS_KEY_ID="XXXXXXXXXXXXXXX" \
AWS_SECRET_ACCESS_KEY="XXXXXXXXXXXXXXXXXX" \
AWS_REGION="xx-xxx-x" \
CACHE_URI="host.docker.internal:6379" \
DB_URI="mongodb://host.docker.internal:27017" \
ENV="dev" \
ADMIN_USER="someuser" \
ADMIN_PASSWORD="somepassword" \
SITE="somewhere.com" \
SESSION_TIMEOUT="86400" \
MONGO_DATABASE="blogdata" \
REDIS_DB="redisdb" \
go run blog.go
```

## Installation Instructions

```bash
go build -o blog blog.go
```

### Docker

```
docker build .
```

## Kubernetes
In the helm directory are two helm charts for the sites I use.  

## Jenkins
There is a sample jenkins file which can be adapted to your tastes

## Customization
There are two template directories which contain the site templates that I manage.  You can customize these for your desired needs.  You can also
copy them and then edit them.  Just make sure your docker build includes the new directory and that you set up the environment configuration
to point to this directory.  You can see how this is accomplished in the Docker file.  




