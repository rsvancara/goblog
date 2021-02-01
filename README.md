# GoBlog - Go Based Photography Blogging Software

This is my take on a simple photography blog platform that I am using for my blog sites. 

Note: This software is a work in progress.  You can see it action here:

https://tryingadventure.com

https://diytinytrailer.com

# Notable features

1.  Supports libvips for image processing
2.  Uses S3 backed storage, but does not expose the s3 or cloudfront backend.  (This may or may not be a good idea)
3.  Uses MongoDB
4.  Uses Redis for storing sessions
5.  Features image management interface to upload, extract Exif information.  
6.  Uses Markdown language for the editing environment, just like you would in github.  The reason for this is that I 
    did not want to incorporate a clunky WSIWYG editor that generates HTML differently every new release.  All text is store 
    as markdown. 
7.  Fast!  The software is written in golang and compiles to a single binary.  Maybe the next version will be in Rust...

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




