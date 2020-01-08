# DaBloog - Go Based Blogging Software

Sample blogging software for testing application deployments

# Instructions

```bash
AWS_PROFILE="xxxx" \
S3_BUCKET="xxxx" \
AWS_ACCESS_KEY_ID="XXXXXXXXXXXXXXX" \
AWS_SECRET_ACCESS_KEY="XXXXXXXXXXXXXXXXXX" \
AWS_REGION="xx-xxx-x" \
CACHE_URI="host.docker.internal:6379" \
DB_URI="mongodb://host.docker.internal:27017" \
go run blog.go
```

## Installation, Instructions

```bash
go build -o blog blog.go
```

### Docker

```
docker build .
```


